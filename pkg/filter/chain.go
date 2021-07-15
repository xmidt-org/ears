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
	"fmt"
	"github.com/xmidt-org/ears/internal/pkg/rtsemconv"
	"github.com/xmidt-org/ears/pkg/tenant"
	"go.opentelemetry.io/otel"

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

func (c *Chain) Filter(e event.Event) []event.Event {
	c.Lock()
	defer c.Unlock()
	// pass event through in case of empty filter chain
	if len(c.filterers) == 0 {
		return []event.Event{e}
	}
	type work struct {
		e event.Event
		f Filterer
		i int // Index of current filterer
	}
	queue := list.New()
	queue.PushBack(work{e: e, f: c.filterers[0], i: 0})
	events := []event.Event{}
	ctx := e.Context()

	tracer := otel.Tracer(rtsemconv.EARSTracerName)
	for elem := queue.Front(); elem != nil; elem = elem.Next() {
		select {
		case <-ctx.Done():
			return nil
		default:
			w := elem.Value.(work)

			_, span := tracer.Start(w.e.Context(), w.f.Name())
			evts := w.f.Filter(w.e)
			span.End()

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
	return events
}

func (c *Chain) Config() interface{} {
	return nil
}

func (c *Chain) Name() string {
	return "filter_chain"
}

func (c *Chain) Plugin() string {
	return "filter_chain"
}

func (c *Chain) Tenant() tenant.Id {
	return tenant.Id{}
}
