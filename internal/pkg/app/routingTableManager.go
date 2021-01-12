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
	"errors"
	"github.com/xmidt-org/ears/internal/pkg/plugin"
	"github.com/xmidt-org/ears/pkg/route"
	"sync"
)

type DefaultRoutingTableManager struct {
	sync.Mutex
	pluginMgr    plugin.Manager
	storageMgr   route.RouteStorer
	liveRouteMap map[string]*route.Route // to hold in-memory references to live routes
}

func NewRoutingTableManager(pluginMgr plugin.Manager, storageMgr route.RouteStorer) RoutingTableManager {
	return &DefaultRoutingTableManager{pluginMgr: pluginMgr, storageMgr: storageMgr, liveRouteMap: make(map[string]*route.Route, 0)}
}

func (r *DefaultRoutingTableManager) RemoveRoute(ctx context.Context, routeId string) error {
	err := r.storageMgr.DeleteRoute(ctx, routeId)
	if err != nil {
		return err
	}
	//TODO: consider idempotency
	liveRoute, ok := r.liveRouteMap[routeId]
	if !ok {
		return errors.New("no live route exists with ID " + routeId)
	}
	r.Lock()
	delete(r.liveRouteMap, routeId)
	r.Unlock()
	err = liveRoute.Stop(ctx)
	if err != nil {
		return err
	}
	return nil
}

func (r *DefaultRoutingTableManager) AddRoute(ctx context.Context, routeConfig *route.Config) error {
	err := r.storageMgr.SetRoute(ctx, *routeConfig)
	if err != nil {
		return err
	}
	//buf, _ := json.Marshal(routeConfig)
	//fmt.Printf("%v\n", string(buf))
	receiver, err := r.pluginMgr.RegisterReceiver(ctx, routeConfig.Receiver.Plugin, routeConfig.Receiver.Name, routeConfig.Receiver.Config)
	if err != nil {
		return err
	}
	sender, err := r.pluginMgr.RegisterSender(ctx, routeConfig.Sender.Plugin, routeConfig.Sender.Name, routeConfig.Sender.Config)
	if err != nil {
		return err
	}
	liveRoute := &route.Route{}
	//TODO: what about hash generation and validation?
	r.Lock()
	r.liveRouteMap[routeConfig.Id] = liveRoute
	r.Unlock()
	go func() {
		//TODO: how to pass in a filter chain?
		liveRoute.Run(ctx, receiver, nil, sender) // run is blocking
		//TODO: what to do when Run() returns?
		//TODO: what to do if Run() returns an error?
	}()
	return nil
}

func (r *DefaultRoutingTableManager) GetRoute(ctx context.Context, routeId string) (*route.Config, error) {
	route, err := r.storageMgr.GetRoute(ctx, routeId)
	//TODO: hpow do we treat a non-existent route? is it error worthy?
	if err != nil {
		return nil, err
	}
	return route, nil
}

func (r *DefaultRoutingTableManager) GetAllRoutes(ctx context.Context) ([]route.Config, error) {
	//TODO: shouldn't the type be []*route.Config?
	routes, err := r.storageMgr.GetAllRoutes(ctx)
	if err != nil {
		return nil, err
	}
	return routes, nil
}
