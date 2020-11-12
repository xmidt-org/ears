/**
 *  Copyright (c) 2020  Comcast Cable Communications Management, LLC
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package route

import (
	"context"
	"sync"
)

type (
	// A FilterChain is a slice of filter plugins
	FilterChain struct {
		Filters []*FilterPlugin `json:"filters"` // optional list of filter plugins that will be applied in order to perform arbitrary filtering and transformation functions
		lock    sync.RWMutex
	}
)

func NewFilterChain(ctx context.Context) *FilterChain {
	fc := new(FilterChain)
	fc.Filters = make([]*FilterPlugin, 0)
	return fc
}

// Initialize sets up filter chain objects and channels based on filter chain config
func (fc *FilterChain) Initialize(ctx context.Context, rte *Route) error {
	var err error
	var eventChannel chan *Event
	rte.lock.Lock()
	defer rte.lock.Unlock()
	if fc.Filters == nil {
		fc.Filters = make([]*FilterPlugin, 0)
	}
	if len(fc.Filters) == 0 {
		// seed an empty filter chain with a pass (aka nop) filter
		fp := new(FilterPlugin)
		fp.Type = FilterTypePass
		fc.Filters = append(fc.Filters, fp)
	}
	if fc.Filters != nil {
		for idx, fp := range fc.Filters {
			fp.filterer, err = NewFilterer(ctx, fp)
			if err != nil {
				return err
			}
			fp.State = PluginStateReady
			fp.Mode = PluginModeFilter
			fp.routes = []*Route{rte}
			fp.lock.Lock()
			fp.done = make(chan bool)
			fp.lock.Unlock()
			if idx == 0 {
				// each filter chain gets its own set of channels
				fp.SetInputChannel(make(chan *Event))
			} else {
				fp.SetInputChannel(eventChannel)
			}
			if idx == len(fc.Filters)-1 {
				// last channel in the filter chain is the (shared) input channel of the output plugin
				fp.SetOutputChannel(rte.Destination.GetInputChannel())
			} else {
				fp.SetOutputChannel(make(chan *Event))
			}
			eventChannel = fp.GetOutputChannel()
			fp.DoAsync(ctx)
		}
	}
	return nil
}

func (fc *FilterChain) Withdraw(ctx context.Context) {
	if fc.Filters != nil {
		for _, f := range fc.Filters {
			f.Close(ctx)
		}
	}
}
