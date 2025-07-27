// SPDX-FileCopyrightText: 2025 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package tablemgr

import (
	"context"
	"sync"
	"sync/atomic"

	"github.com/xmidt-org/ears/pkg/filter"
	"github.com/xmidt-org/ears/pkg/receiver"
	"github.com/xmidt-org/ears/pkg/route"
	"github.com/xmidt-org/ears/pkg/sender"
)

type LiveRouteWrapper struct {
	sync.Mutex
	Route       *route.Route
	Sender      sender.Sender
	Receiver    receiver.Receiver
	FilterChain *filter.Chain
	Config      route.Config
	RefCnt      int32
}

func NewLiveRouteWrapper(routeConfig route.Config) *LiveRouteWrapper {
	lrw := new(LiveRouteWrapper)
	lrw.Config = routeConfig
	atomic.AddInt32(&lrw.RefCnt, 1)
	return lrw
}

func (lrw *LiveRouteWrapper) GetReferenceCount() int {
	return int(lrw.RefCnt)
}

func (lrw *LiveRouteWrapper) AddRouteReference() int {
	atomic.AddInt32(&lrw.RefCnt, 1)
	return int(lrw.RefCnt)
}

func (lrw *LiveRouteWrapper) RemoveRouteReference() int {
	atomic.AddInt32(&lrw.RefCnt, -1)
	return int(lrw.RefCnt)
}

func (lrw *LiveRouteWrapper) Unregister(ctx context.Context, r *DefaultRoutingTableManager) error {
	lrw.Lock()
	defer lrw.Unlock()
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
	tid := lrw.Config.TenantId
	if lrw.Config.FilterChain != nil {
		for _, f := range lrw.Config.FilterChain {
			filter, err := r.pluginMgr.RegisterFilter(ctx, f.Plugin, f.Name, stringify(f.Config), tid)
			if err != nil {
				lrw.Unregister(ctx, r)
				return err
			}
			lrw.FilterChain.Add(filter)
		}
	}
	// set up sender
	lrw.Sender, err = r.pluginMgr.RegisterSender(ctx, lrw.Config.Sender.Plugin, lrw.Config.Sender.Name, stringify(lrw.Config.Sender.Config), tid)
	if err != nil {
		lrw.Unregister(ctx, r)
		return err
	}
	// set up receiver
	lrw.Receiver, err = r.pluginMgr.RegisterReceiver(ctx, lrw.Config.Receiver.Plugin, lrw.Config.Receiver.Name, stringify(lrw.Config.Receiver.Config), tid)
	if err != nil {
		lrw.Unregister(ctx, r)
		return err
	}
	return nil
}
