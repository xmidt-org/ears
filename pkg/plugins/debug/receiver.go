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
	"sync"

	"fmt"
	"time"

	"github.com/goccy/go-yaml"
	"github.com/xmidt-org/ears/pkg/event"
	pkgplugin "github.com/xmidt-org/ears/pkg/plugin"
	"github.com/xmidt-org/ears/pkg/receiver"
)

func (r *Receiver) Receive(next receiver.NextFn) error {
	if r == nil {
		return &pkgplugin.Error{
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
				r.done <- struct{}{}
			}
			r.Unlock()
		}()

		eventsDone := &sync.WaitGroup{}
		eventsDone.Add(*r.config.Rounds)
		for count := *r.config.Rounds; count != 0; {
			select {
			case <-r.done:
				//fmt.Printf("RECEIVER DONE BY ROUTE\n")
				return
			case <-time.After(time.Duration(*r.config.IntervalMs) * time.Millisecond):
				//fmt.Printf("RECEIVER NEW EVENT\n")
				ctx := context.Background()
				e, err := event.NewEventWithAck(ctx, r.config.Payload,
					func() {
						eventsDone.Done()
					}, func(err error) {
						//fmt.Println("Debug error", err.Error())
						eventsDone.Done()
					})
				if err != nil {
					return
				}

				r.Trigger(e)

				if count > 0 {
					count--
				}
			}
		}
		eventsDone.Wait()
	}()

	<-r.done
	return nil
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

	cfg = cfg.WithDefaults()

	err = cfg.Validate()
	if err != nil {
		return nil, err
	}

	r := &Receiver{
		config: cfg,
	}

	r.history = newHistory(*r.config.MaxHistory)

	return r, nil
}

func (r *Receiver) Count() int {
	return r.history.Count()
}

func (r *Receiver) Trigger(e event.Event) {
	// Ensure that `next` can be slow and locking here will not
	// prevent other requests from executing.
	//BW TODO: where is the next slice here?
	r.Lock()
	next := r.next
	//BW
	//fmt.Printf("TRIGGER\n")
	r.Unlock()

	r.history.Add(e)

	next(e)
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
