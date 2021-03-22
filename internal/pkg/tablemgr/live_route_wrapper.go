// Copyright 2020 Comcast Cable Communications Management, LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package tablemgr

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
	RefCnt      int
}

func NewLiveRouteWrapper(routeConfig route.Config) *LiveRouteWrapper {
	lrw := new(LiveRouteWrapper)
	lrw.Config = routeConfig
	lrw.RefCnt++
	return lrw
}

func (lrw *LiveRouteWrapper) GetReferenceCount() int {
	lrw.Lock()
	defer lrw.Unlock()
	return lrw.RefCnt
}

func (lrw *LiveRouteWrapper) AddRouteReference() int {
	lrw.Lock()
	defer lrw.Unlock()
	lrw.RefCnt++
	return lrw.RefCnt
}

func (lrw *LiveRouteWrapper) RemoveRouteReference() int {
	lrw.Lock()
	defer lrw.Unlock()
	lrw.RefCnt--
	return lrw.RefCnt
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
	lrw.Lock()
	defer lrw.Unlock()
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
