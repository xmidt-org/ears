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

	"github.com/goccy/go-yaml"
	"github.com/xmidt-org/ears/pkg/event"
	pkgplugin "github.com/xmidt-org/ears/pkg/plugin"
	"github.com/xmidt-org/ears/pkg/receiver"
)

func (r *Receiver) Receive(ctx context.Context, next receiver.NextFn) error {
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
		for count := r.config.Rounds; count != 0; count-- {
			select {
			case <-ctx.Done():
				return
			case <-time.After(time.Duration(r.config.IntervalMs) * time.Millisecond):
				e, err := event.NewEvent(r.config.Payload)
				if err != nil {
					return
				}

				err = r.Trigger(ctx, e)
				if err != nil {
					return
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

	var cfg ReceiverConfig
	var err error

	switch c := config.(type) {
	case string:
		err = yaml.Unmarshal([]byte(c), &cfg)
	case []byte:
		err = yaml.Unmarshal(c, &cfg)
	case ReceiverConfig:
		cfg = c
	case *ReceiverConfig:
		cfg = *c
	}

	if err != nil {
		return nil, &pkgplugin.InvalidConfigError{
			Err: err,
		}
	}

	r := &Receiver{
		config: *(cfg.WithDefaults()),
	}

	err = r.config.Validate()
	if err != nil {
		return nil, err
	}

	r.history = newHistory(r.config.MaxHistory)

	return r, nil
}

func (r *Receiver) Count() int {
	return r.history.Count()
}

func (r *Receiver) Trigger(ctx context.Context, e event.Event) error {
	r.Lock()
	defer r.Unlock()

	r.history.Add(e)

	return r.next(ctx, e)
}

func (r *Receiver) History() []event.Event {
	history := r.history.History()

	events := make([]event.Event, len(history))

	for i, h := range history {
		if e, ok := h.(event.Event); ok {
			events[i] = e
		}
	}

	return events

}
