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
	"fmt"
	"time"

	"github.com/rs/zerolog/log"
)

//TODO: separate configs from state
//TODO: use interface rather than struct
//TODO: add org id and app id to plugin and consider in hash calculation

type (
	// An EarsPlugin represents an input plugin an output plugin or a filter plugin
	Plugin struct {
		Type       string                 `json:"type"`       // plugin or filter type, e.g. kafka, kds, sqs, webhook, filter
		Version    string                 `json:"version"`    // plugin version
		SOName     string                 `json:"soName"`     // name of shared library file implementing this plugin
		Params     map[string]interface{} `json:"params"`     // plugin specific configuration parameters
		Mode       string                 `json:"mode"`       // plugin mode, one of input, output and filter
		State      string                 `json:"state"`      // plugin operational state including running, stopped, error etc. (filter plugins are always in state running)
		Name       string                 `json:"name"`       // descriptive plugin name
		Encodings  []string               `json:"encodings"`  // list of supported encodings
		EventCount int                    `json:"eventCount"` // number of events that have passed through this plugin
		RouteCount int                    `json:"routeCount"` // number of routes using this plugin
	}
	// IOPlugin represents an input or an output plugin
	IOPlugin struct {
		Plugin
		routes []*RoutingTableEntry // list of routes using this plugin instance as source plugin
	}
	// FilterPlugin represents a filter plugin
	FilterPlugin struct {
		Plugin
		routingTableEntry *RoutingTableEntry // routing table entry this fiter plugin belongs to
		inputChannel      chan *Event        // channel on which this filter receives the next event
		outputChannel     chan *Event        // channel to which this filter forwards this event to
		done              chan bool          // done channel
		filterer          Filterer           // an instance of the appropriate filterer
		// note: if event is filtered it will not be forwarded
		// note: if event is split multiple events will be forwarded
		// note: if output channel is nil, we are at the end of the filter chain and the event is to be delivered to the output plugin of the route
	}
	DebugInputPlugin struct {
		IOPlugin
		IntervalMs  int
		Rounds      int
		Payload     interface{}
		EventQueuer EventQueuer
	}
	DebugOutputPlugin struct {
		IOPlugin
	}
)

func (plgn *Plugin) Hash(ctx context.Context) string {
	str := ""
	// distinguish different plugin types
	str += plgn.Type
	// distinguish different configurations
	if plgn.Params != nil {
		buf, _ := json.Marshal(plgn.Params)
		str += string(buf)
	}
	// distinguish input and output plugins
	str += plgn.Mode
	// optionally distinguish by org and app here as well
	hash := fmt.Sprintf("%x", md5.Sum([]byte(str)))
	return hash
}

func (plgn *Plugin) Validate(ctx context.Context) error {
	return nil
}

func (plgn *Plugin) Initialize(ctx context.Context) error {
	return nil
}

func (plgn *Plugin) String() string {
	buf, _ := json.Marshal(plgn)
	return string(buf)
}

func (dip *DebugInputPlugin) DoAsync(ctx context.Context) {
	done := false
	go func() {
		if dip.EventQueuer == nil {
			log.Error().Msg("no event queue set for debug input plugin " + dip.Hash(ctx))
			return
		}
		if dip.Payload == nil {
			log.Error().Msg("no payload configured for debug input plugin " + dip.Hash(ctx))
			return
		}
		for {
			if dip.Rounds > 0 && dip.EventCount >= dip.Rounds {
				return
			}
			time.Sleep(time.Duration(dip.IntervalMs) * time.Millisecond)
			if done {
				break
			}
			event := NewEvent(ctx, &dip.IOPlugin, dip.Payload)
			log.Debug().Msg("debug input plugin " + dip.Hash(ctx) + " produced event " + fmt.Sprintf("%d", dip.EventCount))
			// place event on buffered event channel
			dip.EventQueuer.AddEvent(ctx, event)
			dip.EventCount++
		}
	}()
}

func (dip *DebugInputPlugin) DoSync(ctx context.Context) {
	if dip.EventQueuer == nil {
		log.Error().Msg("no event queue set for debug input plugin " + dip.Hash(ctx))
		return
	}
	if dip.Payload == nil {
		log.Error().Msg("no payload configured for debug input plugin " + dip.Hash(ctx))
		return
	}
	event := NewEvent(ctx, &dip.IOPlugin, dip.Payload)
	log.Debug().Msg("debug input plugin " + dip.Hash(ctx) + " produced event " + fmt.Sprintf("%d", dip.EventCount))
	// place event on buffered event channel
	dip.EventQueuer.AddEvent(ctx, event)
	dip.EventCount++
}

func (dop *DebugOutputPlugin) DoSync(ctx context.Context, event *Event) error {
	log.Debug().Msg("debug output plugin " + dop.Hash(ctx) + " consumed event " + fmt.Sprintf("%d", dop.EventCount))
	dop.EventCount++
	return nil
}

func (dop *DebugOutputPlugin) DoAsync(ctx context.Context) {
}

func (iop *IOPlugin) DoSync(ctx context.Context, event *Event) error {
	//TODO: improve
	if iop.Mode == PluginModeOutput {
		log.Debug().Msg("debug output plugin " + iop.Hash(ctx) + " consumed event " + fmt.Sprintf("%d", iop.EventCount))
		iop.EventCount++
	}
	return nil
}

func NewInputPlugin(ctx context.Context, rte *RoutingTableEntry) (*IOPlugin, error) {
	switch rte.Source.Type {
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
		dip.Params = rte.Source.Params
		dip.routes = []*RoutingTableEntry{rte}
		dip.RouteCount = 1
		dip.EventQueuer = GetEventQueue(ctx)
		// parse configs and overwrite defaults
		if dip.Params != nil {
			if value, ok := dip.Params["rounds"].(float64); ok {
				dip.Rounds = int(value)
			}
			if value, ok := dip.Params["intervalMS"].(float64); ok {
				dip.IntervalMs = int(value)
			}
			if value, ok := dip.Params["payload"]; ok {
				dip.Payload = value
			}
		}
		// start producing events
		dip.DoAsync(ctx)
		return &dip.IOPlugin, nil
	}
	return nil, &UnknownInputPluginTypeError{rte.Source.Type}
}

func NewOutputPlugin(ctx context.Context, rte *RoutingTableEntry) (*IOPlugin, error) {
	switch rte.Destination.Type {
	case PluginTypeDebug:
		dop := new(DebugOutputPlugin)
		dop.Type = PluginTypeDebug
		dop.Mode = PluginModeOutput
		dop.State = PluginStateReady
		dop.Name = "Debug"
		dop.Params = rte.Destination.Params
		dop.routes = []*RoutingTableEntry{rte}
		dop.RouteCount = 1
		dop.DoAsync(ctx)
		return &dop.IOPlugin, nil
	}
	return nil, &UnknownOutputPluginTypeError{rte.Destination.Type}
}
