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
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/rs/zerolog/log"
)

type (
	// An EarsPlugin represents an input plugin an output plugin or a filter plugin
	Plugin struct {
		Type    string                 `json:"type"` // source plugin type, e.g. kafka, kds, sqs, webhook, filter
		Version string                 `json:"version"`
		Params  map[string]interface{} `json:"params"` // plugin specific configuration parameters
		Mode    string                 `json:"mode"`   // plugin mode, one of input, output and filter
		State   string                 `json:"state"`  // plugin operational state including running, stopped, error etc. (filter plugins are always in state running)
		Name    string                 `json:"name"`   // descriptive plugin name
	}

	InputPlugin struct {
		Plugin
		routes []*RoutingTableEntry // list of routes using this plugin instance as source plugin
	}

	OutputPlugin struct {
		Plugin
		routes []*RoutingTableEntry // list of routes using this plugin instance as destination plugin
	}

	FilterPlugin struct {
		Plugin
		routingTableEntry *RoutingTableEntry // routing table entry this fiter plugin belongs to
		inputChannel      chan *Event        // channel on which this filter receives the next event
		outputChannel     chan *Event        // channel to which this filter forwards this event to
		filterer          Filterer           // an instance of the appropriate filterer
		// note: if event is filtered it will not be forwarded
		// note: if event is split multiple events will be forwarded
		// note: if output channel is nil, we are at the end of the filter chain and the event is to be delivered to the output plugin of the route
	}
)

func (plgn *Plugin) Hash(ctx context.Context) string {
	str := plgn.Type
	if plgn.Params != nil {
		buf, _ := json.Marshal(plgn.Params)
		str += string(buf)
	}
	hash := fmt.Sprintf("%x", md5.Sum([]byte(str)))
	return hash
}

func (plgn *Plugin) Validate(ctx context.Context) error {
	return nil
}

func (plgn *Plugin) Initialize(ctx context.Context) error {
	return nil
}

type (
	DebugInputPlugin struct {
		InputPlugin
		IntervalMs  int
		Rounds      int
		Payload     interface{}
		EventQueuer EventQueuer
	}

	DebugOutputPlugin struct {
		OutputPlugin
	}
)

func (dip *DebugInputPlugin) DoAsync(ctx context.Context) {
	go func() {
		if dip.EventQueuer == nil {
			log.Error().Msg("no event queue set for debug input plugin " + dip.Hash(ctx))
			return
		}
		if dip.Payload == nil {
			log.Error().Msg("no payload configured for debug input plugin " + dip.Hash(ctx))
			return
		}
		cnt := 0
		for {
			if dip.Rounds > 0 && cnt >= dip.Rounds {
				return
			}
			time.Sleep(time.Duration(dip.IntervalMs) * time.Millisecond)
			event := NewEvent(ctx, &dip.InputPlugin, dip.Payload)
			log.Debug().Msg("debug input plugin " + dip.Hash(ctx) + " produced event")
			// place event on buffered event channel
			dip.EventQueuer.AddEvent(ctx, event)
			cnt++
		}
	}()
}

func (dop *DebugOutputPlugin) DoSync(ctx context.Context, event *Event) error {
	log.Debug().Msg("debug output plugin " + dop.Hash(ctx) + " consumed event")
	return nil
}

func (op *OutputPlugin) DoSync(ctx context.Context, event *Event) error {
	log.Debug().Msg("output plugin " + op.Hash(ctx) + " consumed event")
	return nil
}

func NewInputPlugin(ctx context.Context, rte *RoutingTableEntry) (*InputPlugin, error) {
	switch rte.SrcType {
	case PluginTypeDebug:
		dip := new(DebugInputPlugin)
		// initialize with defaults
		dip.Payload = map[string]string{"hello": "world"}
		dip.IntervalMs = 1000
		dip.Rounds = 1
		dip.Type = PluginTypeDebug
		dip.Mode = PluginModeInput
		dip.State = PluginStateReady
		dip.Name = "Debug"
		dip.Params = rte.SrcParams
		dip.routes = []*RoutingTableEntry{rte}
		dip.EventQueuer = GetEventQueue(ctx)
		// parse configs and overwrite defaults
		if rte.SrcParams != nil {
			if value, ok := rte.SrcParams["rounds"].(float64); ok {
				dip.Rounds = int(value)
			}
			if value, ok := rte.SrcParams["intervalMS"].(float64); ok {
				dip.IntervalMs = int(value)
			}
			if value, ok := rte.SrcParams["payload"]; ok {
				dip.Payload = value
			}
		}
		// start producing events
		dip.DoAsync(ctx)
		return &dip.InputPlugin, nil
	}
	return nil, errors.New("unknown input plugin type " + rte.SrcType)
}

func NewOutputPlugin(ctx context.Context, rte *RoutingTableEntry) (*OutputPlugin, error) {
	switch rte.DstType {
	case PluginTypeDebug:
		dop := new(DebugOutputPlugin)
		dop.Type = PluginTypeDebug
		dop.Mode = PluginModeOutput
		dop.State = PluginStateReady
		dop.Name = "Debug"
		dop.Params = rte.DstParams
		dop.routes = []*RoutingTableEntry{rte}
		return &dop.OutputPlugin, nil
	}
	return nil, errors.New("unknown output plugin type " + rte.DstType)
}
