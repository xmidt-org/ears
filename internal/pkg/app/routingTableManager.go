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
	"github.com/xmidt-org/ears/pkg/event"

	"github.com/xmidt-org/ears/internal/pkg/plugin"
	"github.com/xmidt-org/ears/pkg/route"
)

type DefaultRoutingTableManager struct {
	pluginMgr  plugin.Manager
	storageMgr route.RouteStorer
}

func NewRoutingTableManager(plugMgr plugin.Manager, storageMgr route.RouteStorer) RoutingTableManager {
	return &DefaultRoutingTableManager{plugMgr, storageMgr}
}

func (r *DefaultRoutingTableManager) AddRoute(ctx context.Context, route *route.Config) error {
	err := r.storageMgr.SetRoute(ctx, *route)
	if err != nil {
		return err
	}
	receiver, err := r.pluginMgr.RegisterReceiver(ctx, "debug", "myDebugIn", route.Source.Config)
	sender, err := r.pluginMgr.RegisterSender(ctx, "debug", "myDebugOut", route.Destination.Config)
	go func() {
		receiver.Receive(ctx, func(ctx context.Context, e event.Event) error {
			return sender.Send(ctx, e)
		})
	}()
	/*for _, f := range route.FilterChain {
		r.pluginMgr.RegisterFilter(ctx, "match", "myMatch", f.Config)
	}*/
	return nil
}
