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
)

func NewFilterChain(ctx context.Context) *FilterChain {
	fc := new(FilterChain)
	fc.Filters = make([]*FilterPlugin, 0)
	return fc
}

// Initialize sets up filter chain objects and channels based on filter chain config
func (fc *FilterChain) Initialize(ctx context.Context, rte *Route) error {
	//
	// initialize filter chain
	//
	// for now output plugin is not connected via channel but rather via function call
	// input plugin is decoupled from filter chain via simple buffered channel
	// therefore the first filter reads from the buffered event channel and the last filter in the filter chain has a nil output channel
	// we will likely change this in the future
	var err error
	var eventChannel chan *Event
	if fc.Filters == nil {
		fc.Filters = make([]*FilterPlugin, 0)
	}
	if len(fc.Filters) == 0 {
		// seed an empty filter chain with a pass filter
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
			fp.done = make(chan bool)
			if idx == 0 {
				fp.inputChannel = rte.Source.GetOutputChannel()
			} else {
				fp.inputChannel = eventChannel
			}
			fp.outputChannel = make(chan *Event)
			eventChannel = fp.outputChannel
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
