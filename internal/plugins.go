package internal

import (
	"context"
	"crypto/md5"
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

type (
	// An EarsPlugin represents an input plugin an output plugin or a filter plugin
	Plugin struct {
		//Hash         string               `json: "hash"`      // hash over all plugin configurations
		Type    string      `json: "type"` // source plugin type, e.g. kafka, kds, sqs, webhook, filter
		Version string      `json: "version"`
		Params  interface{} `json: "params"` // plugin specific configuration parameters
		Mode    string      `json:"mode"`    // plugin mode, one of input, output and filter
		State   string      `json: "state"`  // plugin operational state including running, stopped, error etc. (filter plugins are always in state running)
		Name    string      `json: "name"`   // descriptive plugin name
	}

	InputPlugin struct {
		Plugin
		Routes []*RoutingTableEntry // list of routes using this plugin instance as source plugin
	}

	OutputPlugin struct {
		Plugin
		Routes []*RoutingTableEntry // list of routes using this plugin instance as destination plugin
	}

	FilterPlugin struct {
		Plugin
		RoutingTableEntry *RoutingTableEntry // routing table entry this fiter plugin belongs to
		InputChannel      chan *Event        // channel on which this filter receives the next event
		OutputChannel     chan *Event        // channel to which this filter forwards this event to
		Filterer          Filterer           // an instance of the appropriate filterer
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
		IntervalMs int
		Payload    interface{}
		eventQueue chan *Event
	}

	DebugOutputPlugin struct {
		OutputPlugin
	}
)

func (dip *DebugInputPlugin) DoAsync(ctx context.Context) {
	go func() {
		if dip.eventQueue == nil {
			return
		}
		if dip.Payload == nil {
			return
		}
		for {
			time.Sleep(time.Duration(dip.IntervalMs) * time.Millisecond)
			event := NewEvent(ctx, &dip.InputPlugin, dip.Payload)
			dip.eventQueue <- event
		}
	}()
}

func (dop *DebugOutputPlugin) DoSync(ctx context.Context, event *Event) error {
	//TODO: print event
	return nil
}

func NewInputPlugin(ctx context.Context, rte *RoutingTableEntry) (*InputPlugin, error) {
	switch rte.SrcType {
	case PluginTypeDebug:
		dip := new(DebugInputPlugin)
		//TODO: pass in config params here
		dip.Payload = map[string]string{"hello": "world"}
		dip.IntervalMs = 1000
		dip.Type = PluginTypeDebug
		dip.Mode = PluginModeInput
		dip.State = PluginStateReady
		dip.Name = "Debug"
		dip.Params = rte.SrcParams
		dip.Routes = []*RoutingTableEntry{rte}
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
		dop.Routes = []*RoutingTableEntry{rte}
		return &dop.OutputPlugin, nil
	}
	return nil, errors.New("unknown output plugin type " + rte.DstType)
}
