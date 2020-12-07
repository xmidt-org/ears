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
	"crypto/md5"
	"encoding/json"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
)

//TODO: separate configs from state
//TODO: add org id and app id to plugin and consider in hash calculation

type (
	// An EarsPlugin represents an input plugin an output plugin or a filter plugin
	Plugin struct {
		Type          string                 `json:"type,omitempty"`       // plugin or filter type, e.g. kafka, kds, sqs, webhook, filter
		Version       string                 `json:"version,omitempty"`    // plugin version
		SOName        string                 `json:"soName,omitempty"`     // name of shared library file implementing this plugin
		Params        map[string]interface{} `json:"params,omitempty"`     // plugin specific configuration parameters
		Mode          string                 `json:"mode,omitempty"`       // plugin mode, one of input, output and filter
		State         string                 `json:"state,omitempty"`      // plugin operational state including running, stopped, error etc. (filter plugins are always in state running)
		Name          string                 `json:"name,omitempty"`       // descriptive plugin name
		Encodings     []string               `json:"encodings,omitempty"`  // list of supported encodings
		EventCount    int                    `json:"eventCount,omitempty"` // number of events that have passed through this plugin
		lastEvent     *Event                 // latest event received
		routes        []*Route               // list of routes using this plugin instance
		inputChannel  chan *Event            // event channel on which plugin receives the next event
		outputChannel chan *Event            // event channel to which plugin forwards current event to
		done          chan bool              // done channel
		filterer      Filterer               // an instance of the appropriate filterer
		lock          sync.RWMutex           // r/w lock
	}
	FilterPlugin struct {
		Plugin
	}
	DebugInputPlugin struct {
		Plugin
		IntervalMs int
		Rounds     int
		Payload    interface{}
	}
	DebugOutputPlugin struct {
		Plugin
	}
)

//
// generic plugin features
//

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
	// distinguish instances by name
	str += "Debug"
	str += plgn.Name
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

func (plgn *Plugin) GetConfig() *Plugin {
	return plgn
}

func (plgn *Plugin) GetInputChannel() chan *Event {
	plgn.lock.RLock()
	defer plgn.lock.RUnlock()
	return plgn.inputChannel
}

func (plgn *Plugin) SetInputChannel(in chan *Event) {
	plgn.lock.Lock()
	defer plgn.lock.Unlock()
	plgn.inputChannel = in
}

func (plgn *Plugin) GetOutputChannel() chan *Event {
	plgn.lock.RLock()
	defer plgn.lock.RUnlock()
	return plgn.outputChannel
}

func (plgn *Plugin) SetOutputChannel(in chan *Event) {
	plgn.lock.Lock()
	defer plgn.lock.Unlock()
	plgn.outputChannel = in
}

func (plgn *Plugin) GetRouteCount() int {
	if plgn.routes != nil {
		return len(plgn.routes)
	}
	return 0
}

func (plgn *Plugin) GetEventCount() int {
	plgn.lock.RLock()
	defer plgn.lock.RUnlock()
	return plgn.EventCount
}

func (plgn *Plugin) GetLastEvent() *Event {
	plgn.lock.RLock()
	defer plgn.lock.RUnlock()
	return plgn.lastEvent
}

func (plgn *Plugin) IncEventCount(inc int) {
	plgn.lock.Lock()
	defer plgn.lock.Unlock()
	plgn.EventCount += inc
}

func (plgn *Plugin) DoSync(ctx context.Context, event *Event) error {
	return new(UnworthyPluginError)
}

func (plgn *Plugin) DoAsync(ctx context.Context) {
}

func (plgn *Plugin) Close(ctx context.Context) {
	log.Ctx(ctx).Debug().Msg("sending done signal for " + plgn.Mode + " " + plgn.Type + " " + plgn.GetConfig().Name + " " + plgn.Hash(ctx))
	until := time.After(1000 * time.Millisecond)
	select {
	case <-until:
		return
	case plgn.done <- true:
	}
}

//
// debug input plugin
//

func (dip *DebugInputPlugin) DoAsync(ctx context.Context) {
	done := false
	go func() {
		<-dip.done
		dip.lock.Lock()
		done = true
		dip.lock.Unlock()
	}()
	go func() {
		if dip.Payload == nil {
			log.Ctx(ctx).Error().Msg("no payload configured for debug input plugin " + dip.GetConfig().Name + " " + dip.Hash(ctx))
			return
		}
		for {
			if dip.Rounds == 0 && dip.IntervalMs == 0 {
				return
			}
			if dip.Rounds > 0 && dip.GetEventCount() >= dip.Rounds {
				return
			}
			time.Sleep(time.Duration(dip.IntervalMs) * time.Millisecond)
			dip.lock.RLock()
			d := done
			dip.lock.RUnlock()
			if d {
				break
			}
			subCtx := context.WithValue(ctx, "foo", "bar")
			event := NewEvent(subCtx, dip, dip.Payload)
			dip.lock.Lock()
			GetIOPluginManager(ctx).lock.RLock()
			// deliver event to each interested route (first filter in chain)
			if dip.routes != nil {
				for _, r := range dip.routes {
					r.lock.RLock()
					if r.FilterChain != nil && len(r.FilterChain.Filters) > 0 {
						//TODO: clone event
						r.FilterChain.Filters[0].GetInputChannel() <- event
						//log.Ctx(ctx).Debug().Msg("debug " + dip.Mode + " " + dip.Type + " plugin " + dip.GetConfig().Name + " " + dip.Hash(ctx) + " pushed event " + fmt.Sprintf("%d", dip.EventCount) + " into filter " + r.FilterChain.Filters[0].Hash(ctx))
					}
					r.lock.RUnlock()
				}
			}
			log.Ctx(ctx).Debug().Msg("debug " + dip.Mode + " " + dip.Type + " plugin " + dip.GetConfig().Name + " " + dip.Hash(ctx) + " produced event " + fmt.Sprintf("%d", dip.EventCount))
			GetIOPluginManager(ctx).lock.RUnlock()
			dip.lock.Unlock()
			dip.IncEventCount(1)
		}
	}()
}

func (dip *DebugInputPlugin) DoSync(ctx context.Context, event *Event) error {
	if dip.Payload == nil {
		return &MissingPluginConfiguratonError{dip.Type, dip.Hash(ctx), PluginModeInput}
	}
	log.Ctx(ctx).Debug().Msg("debug input plugin " + dip.GetConfig().Name + " " + dip.Hash(ctx) + " produced sync event " + fmt.Sprintf("%d", dip.EventCount))
	dip.lock.Lock()
	GetIOPluginManager(ctx).lock.RLock()
	// deliver event to each interested route (first filter in chain)
	if dip.routes != nil {
		for _, r := range dip.routes {
			r.lock.RLock()
			if r.FilterChain != nil && len(r.FilterChain.Filters) > 0 {
				//TODO: clone event
				r.FilterChain.Filters[0].GetInputChannel() <- event
			}
			r.lock.RUnlock()
		}
	}
	GetIOPluginManager(ctx).lock.RUnlock()
	dip.lock.Unlock()
	dip.IncEventCount(1)
	return nil
}

//
// debug output plugin
//

func (dop *DebugOutputPlugin) DoSync(ctx context.Context, event *Event) error {
	log.Ctx(ctx).Debug().Msg(dop.Mode + " " + dop.Type + " plugin " + dop.GetConfig().Name + " " + dop.Hash(ctx) + " consumed event " + strconv.Itoa(dop.GetEventCount()))
	dop.IncEventCount(1)
	dop.lock.Lock()
	dop.lastEvent = event
	dop.lock.Unlock()
	return nil
}

func (dop *DebugOutputPlugin) DoAsync(ctx context.Context) {
	go func() {
		if dop.GetInputChannel() == nil {
			return
		}
		for {
			select {
			case evt := <-dop.GetInputChannel():
				log.Ctx(ctx).Debug().Msg(dop.Mode + " " + dop.Type + " plugin " + dop.GetConfig().Name + " " + dop.Hash(ctx) + " consumed event " + strconv.Itoa(dop.GetEventCount()))
				dop.IncEventCount(1)
				dop.lock.Lock()
				dop.lastEvent = evt
				dop.lock.Unlock()
			case <-dop.done:
				log.Ctx(ctx).Debug().Msg(dop.Mode + " " + dop.Type + " plugin " + dop.GetConfig().Name + " " + dop.Hash(ctx) + " done")
				return
			}
		}
	}()
}

//
// factory function for input plugin creation
//

func NewInputPlugin(ctx context.Context, rte *Route) (Pluginer, error) {
	pc := rte.Source.GetConfig()
	switch pc.Type {
	case PluginTypeDebug:
		dip := new(DebugInputPlugin)
		// initialize with defaults
		dip.Payload = map[string]string{"hello": "world"}
		dip.IntervalMs = 1000
		dip.Rounds = 1
		dip.Type = PluginTypeDebug
		dip.Mode = PluginModeInput
		dip.State = PluginStateReady
		//dip.Name = "Debug"
		dip.Name = rte.Source.GetConfig().Name
		dip.Params = pc.Params
		dip.routes = []*Route{rte}
		// each filter has its own channel
		//dip.SetOutputChannel(make(chan *Event))
		dip.done = make(chan bool)
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
		return dip, nil
	}
	return nil, &UnknownPluginTypeError{pc.Type}
}

//
// factory function for output plugin creation
//

func NewOutputPlugin(ctx context.Context, rte *Route) (Pluginer, error) {
	pc := rte.Destination.GetConfig()
	switch pc.Type {
	case PluginTypeDebug:
		dop := new(DebugOutputPlugin)
		dop.Type = PluginTypeDebug
		dop.Mode = PluginModeOutput
		dop.State = PluginStateReady
		//dop.Name = "Debug"
		dop.Name = rte.Destination.GetConfig().Name
		dop.Params = pc.Params
		dop.routes = []*Route{rte}
		dop.done = make(chan bool)
		// the input channel of an output plugin may be shared by multiple routes
		dop.SetInputChannel(make(chan *Event))
		dop.DoAsync(ctx)
		return dop, nil
	}
	return nil, &UnknownPluginTypeError{pc.Type}
}
