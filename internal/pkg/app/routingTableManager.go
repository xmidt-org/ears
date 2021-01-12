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
)

type DefaultRoutingTableManager struct {
	pluginMgr  plugin.Manager
	storageMgr route.RouteStorer
	routeMap   map[string]*route.Route // to hold in-memory references to live routes
}

func NewRoutingTableManager(plugMgr plugin.Manager, storageMgr route.RouteStorer) RoutingTableManager {
	return &DefaultRoutingTableManager{plugMgr, storageMgr, make(map[string]*route.Route, 0)}
}

func (r *DefaultRoutingTableManager) RemoveRoute(ctx context.Context, routeId string) error {
	err := r.storageMgr.DeleteRoute(ctx, routeId)
	if err != nil {
		return err
	}
	//TODO: consider idempotency
	//TODO: locks
	liveRoute, ok := r.routeMap[routeId]
	if !ok {
		return errors.New("no live route exists with ID " + routeId)
	}
	err = liveRoute.Stop(ctx)
	if err != nil {
		return err
	}
	return nil
}

func (r *DefaultRoutingTableManager) AddRoute(ctx context.Context, routeConfig *route.Config) error {
	err := r.storageMgr.SetRoute(ctx, *routeConfig)
	//TODO: locks
	if err != nil {
		return err
	}
	receiver, err := r.pluginMgr.RegisterReceiver(ctx, "debug", "myDebugIn", routeConfig.Receiver.Config)
	if err != nil {
		return err
	}
	sender, err := r.pluginMgr.RegisterSender(ctx, "debug", "myDebugOut", routeConfig.Sender.Config)
	if err != nil {
		return err
	}
	route := &route.Route{}
	//TODO: what about hash generation and validation?
	r.routeMap[routeConfig.Id] = route
	go func() {
		//TODO: how to pass in a filter chain?
		route.Run(ctx, receiver, nil, sender) // run is blocking
		//TODO: what to do when Run() returns?
		//TODO: what to do if Run() returns an error?
	}()
	return nil
}
