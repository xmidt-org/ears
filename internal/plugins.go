package internal

import (
	"context"
	"errors"
	"time"
)

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
