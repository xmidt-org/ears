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

package route

import (
	"context"
	"fmt"

	"github.com/xmidt-org/ears/pkg/event"
	"github.com/xmidt-org/ears/pkg/filter"
	"github.com/xmidt-org/ears/pkg/receiver"
	"github.com/xmidt-org/ears/pkg/sender"
	"golang.org/x/sync/errgroup"
)

func (rte *Route) Run(ctx context.Context, r receiver.Receiver, f filter.Filterer, s sender.Sender) error {
	if r == nil {
		return &InvalidRouteError{
			Err: fmt.Errorf("receiver cannot be nil"),
		}
	}

	if s == nil {
		return &InvalidRouteError{
			Err: fmt.Errorf("sender cannot be nil"),
		}
	}

	rte.Lock()
	rte.r = r
	rte.f = f
	rte.s = s
	rte.Unlock()

	var next receiver.NextFn
	if f == nil {
		next = s.Send
	} else {
		next = func(ctx context.Context, e event.Event) error {
			events, err := f.Filter(ctx, e)
			if err != nil {
				return err
			}

			err = fanOut(ctx, events, s.Send)
			if err != nil {
				return err
			}

			return nil

		}
	}

	// TODO:  Deal with errors properly
	return rte.r.Receive(ctx, next)

}

func (rte *Route) Stop(ctx context.Context) error {
	rte.Lock()
	defer rte.Unlock()
	if rte.r == nil {
		return nil
	}

	return rte.r.StopReceiving(ctx)
}

func fanOut(ctx context.Context, events []event.Event, next receiver.NextFn) error {
	if next == nil {
		return &InvalidRouteError{
			Err: fmt.Errorf("next cannot be nil"),
		}
	}

	if len(events) == 0 {
		return nil
	}

	g, ctx := errgroup.WithContext(ctx)

	for _, e := range events {
		g.Go(func() error {
			return next(ctx, e)
		})
	}

	return g.Wait()

}
