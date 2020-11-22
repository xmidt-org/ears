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

package chain

import (
	"context"
	"fmt"

	"github.com/xmidt-org/ears/pkg/event"
	"github.com/xmidt-org/ears/pkg/filter"
	"golang.org/x/sync/errgroup"
)

var _ FiltererChain = (*Chain)(nil)

func (c *Chain) Add(f filter.Filterer) error {
	if f == nil {
		return &InvalidArgumentError{
			Err: fmt.Errorf("filter cannot be nil"),
		}
	}

	c.Lock()
	defer c.Unlock()
	if c.filterers == nil {
		c.filterers = make([]filter.Filterer, 1)
	}

	c.filterers = append(c.filterers, f)

	return nil
}

func (c *Chain) Filterers() []filter.Filterer {
	c.Lock()
	defer c.Unlock()
	fs := []filter.Filterer{}
	fs = append(fs, c.filterers...)
	return fs
}

func (c *Chain) Filter(ctx context.Context, e event.Event) ([]event.Event, error) {

	c.Lock()
	defer c.Unlock()

	eventCh := make(chan event.Event)

	type work struct {
		e event.Event
		f filter.Filterer
		i int // Index of current filterer
	}

	g, ctx := errgroup.WithContext(ctx)
	wrk := make(chan work, len(c.filterers))

	g.Go(func() error {
		select {
		case wrk <- work{e: e, f: c.filterers[0], i: 0}:
		case <-ctx.Done():
			return ctx.Err()
		}
		return nil
	})

	g.Go(func() error {
		select {
		case w := <-wrk:
			evts, err := w.f.Filter(ctx, w.e)
			if err != nil {
				return err
			}

			next := w.i + 1
			if next < len(c.filterers) {
				for _, e := range evts {
					select {
					case wrk <- work{e: e, f: c.filterers[next], i: next}:
					case <-ctx.Done():
						return ctx.Err()
					}
				}
			} else {
				select {
				case eventCh <- e:
				case <-ctx.Done():
					return ctx.Err()
				}

			}

		case <-ctx.Done():
			return ctx.Err()
		}
		return nil
	})

	go func() {
		g.Wait()
		close(eventCh)
	}()

	events := []event.Event{}
	for e := range eventCh {
		events = append(events, e)

	}

	if err := g.Wait(); err != nil {
		return nil, err
	}

	return events, nil

}
