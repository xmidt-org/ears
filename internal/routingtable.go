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
	if rte.FilterChain != nil {
		for idx, fp := range rte.FilterChain {
			var err error
			fp.Filterer, err = NewFilterer(ctx, fp)
			fp.State = PluginStateReady
			fp.Mode = PluginModeFilter
			fp.RoutingTableEntry = rte
			if err != nil {
				return err
			}
			if idx == 0 {
				fp.InputChannel = GetEventQueue(ctx).GetChannel(ctx)
			} else {
				fp.InputChannel = eventChannel
			}
			if idx < len(rte.FilterChain)-1 {
				fp.OutputChannel = make(chan *Event)
				eventChannel = fp.OutputChannel
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

func (fp *FilterPlugin) DoSync(ctx context.Context, event *Event) error {
	log.Debug().Msg(fp.Type + " filter " + fp.Hash(ctx) + " passed")
	filteredEvents, err := fp.Filterer.Filter(ctx, event)
	if err != nil {
		return err
	}
	for _, e := range filteredEvents {
		if fp.OutputChannel != nil {
			fp.OutputChannel <- e
		} else {
			fp.RoutingTableEntry.Destination.DoSync(ctx, e)
		}
	}
	return nil
}

func (fp *FilterPlugin) DoAsync(ctx context.Context) {
	go func() {
		if fp.InputChannel == nil {
			return
		}
		for {
			inputEvent := <-fp.InputChannel
			log.Debug().Msg(fp.Type + " filter " + fp.Hash(ctx) + " passed")
			filteredEvents, err := fp.Filterer.Filter(ctx, inputEvent)
			if err != nil {
				return
			}
			for _, e := range filteredEvents {
				if fp.OutputChannel != nil {
					fp.OutputChannel <- e
				} else {
					fp.RoutingTableEntry.Destination.DoSync(ctx, e)
				}
			}
		}
	}()
}
