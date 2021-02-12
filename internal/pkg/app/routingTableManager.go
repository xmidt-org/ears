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

package app

import (
	"context"
	"encoding/json"
	"github.com/rs/zerolog/log"
	"github.com/xmidt-org/ears/internal/pkg/plugin"
	"github.com/xmidt-org/ears/pkg/filter"
	"github.com/xmidt-org/ears/pkg/receiver"
	"github.com/xmidt-org/ears/pkg/route"
	"github.com/xmidt-org/ears/pkg/sender"
	"sync"
)

type DefaultRoutingTableManager struct {
	sync.Mutex
	pluginMgr    plugin.Manager
	storageMgr   route.RouteStorer
	liveRouteMap map[string]*LiveRouteWrapper // in-memory references to live routes
}

type LiveRouteWrapper struct {
	Route       *route.Route
	Sender      sender.Sender
	Receiver    receiver.Receiver
	FilterChain *filter.Chain
}

func NewRoutingTableManager(pluginMgr plugin.Manager, storageMgr route.RouteStorer) RoutingTableManager {
	return &DefaultRoutingTableManager{pluginMgr: pluginMgr, storageMgr: storageMgr, liveRouteMap: make(map[string]*LiveRouteWrapper, 0)}
}

func (r *DefaultRoutingTableManager) RemoveRoute(ctx context.Context, routeId string) error {
	err := r.storageMgr.DeleteRoute(ctx, routeId)
	if err != nil {
		return err
	}
	liveRoute, ok := r.liveRouteMap[routeId]
	if !ok {
		// no error to make this idempotent
		//return errors.New("no live route exists with ID " + routeId)
	} else {
		r.Lock()
		delete(r.liveRouteMap, routeId)
		r.Unlock()
		err = liveRoute.Route.Stop(ctx)
		if err != nil {
			return err
		}
		// should we really unregister sender, receiver and filter chain here or should this be done by route.Stop()?
		err = r.pluginMgr.UnregisterSender(ctx, liveRoute.Sender)
		if err != nil {
			return err
		}
		err = r.pluginMgr.UnregisterReceiver(ctx, liveRoute.Receiver)
		if err != nil {
			return err
		}
		if liveRoute.FilterChain != nil {
			for _, filter := range liveRoute.FilterChain.Filterers() {
				err = r.pluginMgr.UnregisterFilter(ctx, filter)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func stringify(data interface{}) string {
	if data == nil {
		return ""
	}
	buf, err := json.Marshal(data)
	if err != nil {
		return ""
	}
	return string(buf)
}

func (r *DefaultRoutingTableManager) AddRoute(ctx context.Context, routeConfig *route.Config) error {
	ctx = context.Background()
	// use hashed ID if none is provided - this ID will be returned by the AddRoute REST API
	if routeConfig.Id == "" {
		routeConfig.Id = routeConfig.Hash(ctx)
	}
	err := routeConfig.Validate(ctx)
	if err != nil {
		return err
	}
	// currently storage layer handles created and updated timestamps
	err = r.storageMgr.SetRoute(ctx, *routeConfig)
	if err != nil {
		return err
	}
	// set up sender
	sender, err := r.pluginMgr.RegisterSender(ctx, routeConfig.Sender.Plugin, routeConfig.Sender.Name, stringify(routeConfig.Sender.Config))
	if err != nil {
		return err
	}
	// set up receiver
	receiver, err := r.pluginMgr.RegisterReceiver(ctx, routeConfig.Receiver.Plugin, routeConfig.Receiver.Name, stringify(routeConfig.Receiver.Config))
	if err != nil {
		return err
	}
	// set up filter chain
	filterChain := &filter.Chain{}
	if routeConfig.FilterChain != nil {
		for _, f := range routeConfig.FilterChain {
			filter, err := r.pluginMgr.RegisterFilter(ctx, f.Plugin, f.Name, stringify(f.Config))
			filterChain.Add(filter)
			if err != nil {
				return err
			}
		}
	}
	// create live route
	liveRoute := &route.Route{}
	r.Lock()
	r.liveRouteMap[routeConfig.Id] = &LiveRouteWrapper{liveRoute, sender, receiver, filterChain}
	r.Unlock()
	go func() {
		sctx := context.Background()
		//TODO: use application context here? see issue #51
		err = liveRoute.Run(sctx, receiver, filterChain, sender) // run is blocking
		if err != nil {
			log.Ctx(ctx).Error().Str("op", "AddRoute").Msg(err.Error())
		}
	}()
	return nil
}

func (r *DefaultRoutingTableManager) GetRoute(ctx context.Context, routeId string) (*route.Config, error) {
	route, err := r.storageMgr.GetRoute(ctx, routeId)
	if err != nil {
		return nil, err
	}
	return &route, nil
}

func (r *DefaultRoutingTableManager) GetAllRoutes(ctx context.Context) ([]route.Config, error) {
	//TODO: shouldn't the type be []*route.Config?
	routes, err := r.storageMgr.GetAllRoutes(ctx)
	if err != nil {
		return nil, err
	}
	return routes, nil
}

func (r *DefaultRoutingTableManager) GetAllSenders(ctx context.Context) (map[string]sender.Sender, error) {
	senders := r.pluginMgr.Senders()
	return senders, nil
}

func (r *DefaultRoutingTableManager) GetAllReceivers(ctx context.Context) (map[string]receiver.Receiver, error) {
	receivers := r.pluginMgr.Receivers()
	return receivers, nil
}

func (r *DefaultRoutingTableManager) GetAllFilters(ctx context.Context) (map[string]filter.Filterer, error) {
	filterers := r.pluginMgr.Filters()
	return filterers, nil
}
