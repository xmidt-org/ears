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

package filter

import (
	"container/list"
	"context"
	"fmt"

	"github.com/xmidt-org/ears/pkg/event"
)

func (c *Chain) Add(f Filterer) error {
	if f == nil {
		return &InvalidArgumentError{
			Err: fmt.Errorf("filter cannot be nil"),
		}
	}

	c.Lock()
	defer c.Unlock()
	if c.filterers == nil {
		c.filterers = []Filterer{}
	}

	c.filterers = append(c.filterers, f)

	return nil
}

func (c *Chain) Filterers() []Filterer {
	c.Lock()
	defer c.Unlock()
	fs := []Filterer{}
	fs = append(fs, c.filterers...)
	return fs
}

func (c *Chain) Filter(ctx context.Context, e event.Event) ([]event.Event, error) {

	c.Lock()
	defer c.Unlock()

	type work struct {
		e event.Event
		f Filterer
		i int // Index of current filterer
	}

	queue := list.New()
	queue.PushBack(work{e: e, f: c.filterers[0], i: 0})

	events := []event.Event{}

	for elem := queue.Front(); elem != nil; elem = elem.Next() {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
			w := elem.Value.(work)
			evts, err := w.f.Filter(ctx, w.e)
			// TODO: return errors through metrics
			if err != nil {
				return nil, err
			}

			next := w.i + 1
			if next < len(c.filterers) {
				for _, e := range evts {
					queue.PushBack(work{e: e, f: c.filterers[next], i: next})
				}
			} else {
				events = append(events, evts...)
			}
		}
	}

	return events, nil
}
