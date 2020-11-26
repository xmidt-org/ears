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
	"github.com/xmidt-org/ears/pkg/receiver"
)

// var _ receiver.NewReceiverer = (*Receiver)(nil)
var _ receiver.Receiver = (*Receiver)(nil)

func NewReceiver(config interface{}) (receiver.Receiver, error) {
	return &Receiver{
		IntervalMs: 100,
		Rounds:     4,
		Payload:    "debug message",
	}, nil
}

func (r *Receiver) Receive(ctx context.Context, next receiver.NextFn) error {
	if next == nil {
		return &receiver.InvalidConfigError{
			Err: fmt.Errorf("next cannot be nil"),
		}
	}

	r.done = make(chan struct{})

	go func() {
		defer func() {
			r.Lock()
			if r.done != nil {
				close(r.done)
			}
			r.Unlock()
		}()

		for count := 0; count != r.Rounds; count++ {
			select {
			case <-ctx.Done():
				return
			case <-time.After(time.Duration(r.IntervalMs) * time.Millisecond):
				e, err := event.NewEvent(r.Payload)
				if err != nil {
					return
				}

				err = next(ctx, e)
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
