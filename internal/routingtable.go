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

package internal

import (
	"context"
	"crypto/md5"
	"fmt"

	"github.com/rs/zerolog/log"
)

func (rte *RoutingTableEntry) Hash(ctx context.Context) string {
	str := rte.Source.Hash(ctx) + rte.Destination.Hash(ctx)
	if rte.FilterChain != nil {
		for _, filter := range rte.FilterChain {
			str += filter.Hash(ctx)
		}
	}
	hash := fmt.Sprintf("%x", md5.Sum([]byte(str)))
	return hash
}

func (rte *RoutingTableEntry) Validate(ctx context.Context) error {
	return nil
}

func (rte *RoutingTableEntry) Initialize(ctx context.Context) error {
	//
	// initialize input plugin
	//
	var err error
	rte.Source, err = NewInputPlugin(ctx, rte)
	if err != nil {
		return err
	}
	log.Debug().Msg("init")
	//
	// initialize filter chain
	//
	// for now output plugin is not connected via channel but rather via function call
	// input plugin is decoupled from filter chain via simple buffered channel
	// therefore the first filter reads from the buffered event channel and the last filter in the filter chain has a nil output channel
	// we will likely change this in the future
	var eventChannel chan *Event
	if rte.FilterChain == nil {
		rte.FilterChain = make([]*FilterPlugin, 0)
	}
	if len(rte.FilterChain) == 0 {
		// seed an empty filter chain with a pass filter
		fp := new(FilterPlugin)
		fp.Type = FilterTypePass
		fp.filterer, err = NewFilterer(ctx, fp)
		rte.FilterChain = append(rte.FilterChain, fp)
	}
	if rte.FilterChain != nil {
		for idx, fp := range rte.FilterChain {
			fp.filterer, err = NewFilterer(ctx, fp)
			if err != nil {
				return err
			}
			fp.State = PluginStateReady
			fp.Mode = PluginModeFilter
			fp.routingTableEntry = rte
			if idx == 0 {
				fp.inputChannel = GetEventQueue(ctx).GetChannel(ctx)
			} else {
				fp.inputChannel = eventChannel
			}
			if idx < len(rte.FilterChain)-1 {
				fp.outputChannel = make(chan *Event)
				eventChannel = fp.outputChannel
			}
			fp.DoAsync(ctx)
		}
	}
	//
	// initialize output plugin
	//
	rte.Destination, err = NewOutputPlugin(ctx, rte)
	if err != nil {
		return err
	}
	return nil
}
