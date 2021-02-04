// Copyright 2020 Comcast Cable Communications Management, LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package debug

import (
	"context"

	"fmt"
	"time"

	"github.com/xmidt-org/ears/pkg/event"
	"github.com/xmidt-org/ears/pkg/ledger"
	"github.com/xmidt-org/ears/pkg/plugin"
	"github.com/xmidt-org/ears/pkg/receiver"
)

func (r *Receiver) Receive(ctx context.Context, next receiver.NextFn) error {
	if r == nil {
		return &plugin.Error{
			Err: fmt.Errorf("Receive called on <nil> pointer"),
		}
	}

	if next == nil {
		return &receiver.InvalidConfigError{
			Err: fmt.Errorf("next cannot be nil"),
		}
	}

	r.Lock()
	r.done = make(chan struct{})
	r.next = next
	r.Unlock()

	go func() {
		defer func() {
			r.Lock()
			if r.done != nil {
				close(r.done)
			}
			r.Unlock()
		}()

		// NOTE:  If rounds < 0, this messaging will continue
		// until the context is cancelled.
		for count := *r.config.Rounds; count != 0; {
			select {
			case <-ctx.Done():
				return
			case <-r.done:
				return
			case <-time.After(time.Duration(*r.config.IntervalMs) * time.Millisecond):
				e, err := event.NewEvent(r.config.Payload)
				if err != nil {
					return
				}

				// TODO:  Determine context propagation lifecycle
				//   * https://github.com/xmidt-org/ears/issues/51
				//
				// NOTES: Discussed that this behavior could be determined
				//  by a configuration value
				err = r.Trigger(ctx, e)
				if err != nil {
					return
				}

				if count > 0 {
					count--
				}
			}
		}
	}()

	<-r.done

	return ctx.Err()
}

func (r *Receiver) StopReceiving(ctx context.Context) error {
	r.Lock()
	defer r.Unlock()

	if r.done != nil {
		close(r.done)
	}

	return nil
}

func NewReceiver(config interface{}) (receiver.Receiver, error) {
	cfg, err := NewConfig(config)
	if err != nil {
		return nil, &plugin.InvalidConfigError{
			Err: err,
		}
	}

	cfg = cfg.WithDefaults()

	err = cfg.Validate()
	if err != nil {
		return nil, err
	}

	r := &Receiver{
		config: cfg,
	}

	r.history = ledger.NewLimitedRecorder(*r.config.MaxHistory)

	return r, nil
}

func (r *Receiver) Count() int {
	return r.history.Count()
}

func (r *Receiver) Trigger(ctx context.Context, e event.Event) error {
	// Ensure that `next` can be slow and locking here will not
	// prevent other requests from executing.
	r.Lock()
	next := r.next
	r.Unlock()

	r.history.Add(e)

	return next(ctx, e)
}

func (r *Receiver) Records() []event.Event {
	history, err := r.history.Records()
	if err != nil {
		return []event.Event{}
	}

	events := make([]event.Event, len(history))

	for i, h := range history {
		if e, ok := h.(event.Event); ok {
			events[i] = e
		}
	}

	return events

}
