package app

import (
	"context"
	"github.com/xmidt-org/ears/pkg/filter"
	"github.com/xmidt-org/ears/pkg/receiver"
	"github.com/xmidt-org/ears/pkg/route"
	"github.com/xmidt-org/ears/pkg/sender"
	"sync"
)

type LiveRouteWrapper struct {
	sync.Mutex
	Route       *route.Route
	Sender      sender.Sender
	Receiver    receiver.Receiver
	FilterChain *filter.Chain
	Config      route.Config
	RouteIdMap  map[string]struct{} // map of route IDs who share this route - a simple counter would cause issues with idempotency of AddRoute / RemoveRoute APIs
}

func NewLiveRouteWrapper(routeConfig route.Config) *LiveRouteWrapper {
	lrw := new(LiveRouteWrapper)
	lrw.Config = routeConfig
	lrw.Lock()
	defer lrw.Unlock()
	lrw.RouteIdMap = make(map[string]struct{})
	lrw.RouteIdMap[routeConfig.Id] = struct{}{}
	return lrw
}

func (lrw *LiveRouteWrapper) AddRouteId(routeId string) int {
	lrw.Lock()
	defer lrw.Unlock()
	lrw.RouteIdMap[routeId] = struct{}{}
	return len(lrw.RouteIdMap)
}

func (lrw *LiveRouteWrapper) RemoveRouteId(routeId string) int {
	lrw.Lock()
	defer lrw.Unlock()
	delete(lrw.RouteIdMap, routeId)
	return len(lrw.RouteIdMap)
}

func (lrw *LiveRouteWrapper) Unregister(ctx context.Context, r *DefaultRoutingTableManager) error {
	var e, err error
	if lrw.Receiver != nil {
		err = r.pluginMgr.UnregisterReceiver(ctx, lrw.Receiver)
		if err != nil {
			e = err
		}
	}
	if lrw.Sender != nil {
		err = r.pluginMgr.UnregisterSender(ctx, lrw.Sender)
		if err != nil {
			e = err
		}
	}
	if lrw.FilterChain != nil {
		for _, filter := range lrw.FilterChain.Filterers() {
			err = r.pluginMgr.UnregisterFilter(ctx, filter)
			if err != nil {
				e = err
			}
		}
	}
	return e
}
func (lrw *LiveRouteWrapper) Register(ctx context.Context, r *DefaultRoutingTableManager) error {
	var err error
	lrw.FilterChain = &filter.Chain{}
	if lrw.Config.FilterChain != nil {
		for _, f := range lrw.Config.FilterChain {
			filter, err := r.pluginMgr.RegisterFilter(ctx, f.Plugin, f.Name, stringify(f.Config))
			if err != nil {
				lrw.Unregister(ctx, r)
				return err
			}
			lrw.FilterChain.Add(filter)
		}
	}
	// set up sender
	lrw.Sender, err = r.pluginMgr.RegisterSender(ctx, lrw.Config.Sender.Plugin, lrw.Config.Sender.Name, stringify(lrw.Config.Sender.Config))
	if err != nil {
		lrw.Unregister(ctx, r)
		return err
	}
	// set up receiver
	lrw.Receiver, err = r.pluginMgr.RegisterReceiver(ctx, lrw.Config.Receiver.Plugin, lrw.Config.Receiver.Name, stringify(lrw.Config.Receiver.Config))
	if err != nil {
		lrw.Unregister(ctx, r)
		return err
	}
	return nil
}
