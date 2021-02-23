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
	"errors"
	"github.com/rs/zerolog"
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
	rtSyncer     RoutingTableSyncer
	liveRouteMap map[string]*LiveRouteWrapper // in-memory references to live routes
	logger       *zerolog.Logger
	config       Config
}

type LiveRouteWrapper struct {
	Route       *route.Route
	Sender      sender.Sender
	Receiver    receiver.Receiver
	FilterChain *filter.Chain
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

func NewRoutingTableManager(pluginMgr plugin.Manager, storageMgr route.RouteStorer, logger *zerolog.Logger, config Config) RoutingTableManager {
	rtm := &DefaultRoutingTableManager{pluginMgr: pluginMgr, storageMgr: storageMgr, liveRouteMap: make(map[string]*LiveRouteWrapper, 0), logger: logger, config: config}
	rtm.rtSyncer = NewRedisTableSyncer(rtm, logger, config)
	return rtm
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

func (r *DefaultRoutingTableManager) unregisterAndStopRoute(ctx context.Context, routeId string) error {
	var err error
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
			r.logger.Error().Str("op", "SyncRouteRemoved").Msg(err.Error())
			//return err
		}
		err = liveRoute.Unregister(ctx, r)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *DefaultRoutingTableManager) registerAndRunRoute(ctx context.Context, routeConfig *route.Config) error {
	var err error
	var lrw LiveRouteWrapper
	// set up filter chain
	lrw.FilterChain = &filter.Chain{}
	if routeConfig.FilterChain != nil {
		for _, f := range routeConfig.FilterChain {
			filter, err := r.pluginMgr.RegisterFilter(ctx, f.Plugin, f.Name, stringify(f.Config))
			if err != nil {
				lrw.Unregister(ctx, r)
				return err
			}
			lrw.FilterChain.Add(filter)
		}
	}
	// set up sender
	lrw.Sender, err = r.pluginMgr.RegisterSender(ctx, routeConfig.Sender.Plugin, routeConfig.Sender.Name, stringify(routeConfig.Sender.Config))
	if err != nil {
		lrw.Unregister(ctx, r)
		return err
	}
	// set up receiver
	lrw.Receiver, err = r.pluginMgr.RegisterReceiver(ctx, routeConfig.Receiver.Plugin, routeConfig.Receiver.Name, stringify(routeConfig.Receiver.Config))
	if err != nil {
		lrw.Unregister(ctx, r)
		return err
	}
	// create live route
	lrw.Route = &route.Route{}
	r.Lock()
	r.liveRouteMap[routeConfig.Id] = &lrw
	r.Unlock()
	go func() {
		sctx := context.Background()
		//TODO: use application context here? see issue #51
		err = lrw.Route.Run(sctx, lrw.Receiver, lrw.FilterChain, lrw.Sender) // run is blocking
		if err != nil {
			r.logger.Error().Str("op", "AddRoute").Msg(err.Error())
		}
	}()
	return nil
}

func (r *DefaultRoutingTableManager) RemoveRoute(ctx context.Context, routeId string) error {
	err := r.storageMgr.DeleteRoute(ctx, routeId)
	if err != nil {
		return err
	}
	err = r.rtSyncer.PublishSyncRequest(ctx, routeId, false)
	if err != nil {
		r.logger.Error().Str("op", "RemoveRoute").Msg(err.Error())
	}
	return r.unregisterAndStopRoute(ctx, routeId)
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
	err = r.rtSyncer.PublishSyncRequest(ctx, routeConfig.Id, true)
	if err != nil {
		r.logger.Error().Str("op", "AddRoute").Msg(err.Error())
	}
	return r.registerAndRunRoute(ctx, routeConfig)
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

func (r *DefaultRoutingTableManager) SyncRouteAdded(ctx context.Context, routeId string) error {
	routeConfig, err := r.storageMgr.GetRoute(ctx, routeId)
	if err != nil {
		return err
	}
	return r.registerAndRunRoute(ctx, &routeConfig)
}

func (r *DefaultRoutingTableManager) SyncRouteRemoved(ctx context.Context, routeId string) error {
	return r.unregisterAndStopRoute(ctx, routeId)
}

func (r *DefaultRoutingTableManager) SyncAllRoutes(ctx context.Context) error {
	if len(r.liveRouteMap) > 0 {
		//TODO: or should we unregister all running routes here?
		return errors.New("live route map must be empty for sync all")
	}
	routeConfigs, err := r.storageMgr.GetAllRoutes(ctx)
	if err != nil {
		return err
	}
	//TODO: should this be locked in any way?
	for _, routeConfig := range routeConfigs {
		err = r.registerAndRunRoute(ctx, &routeConfig)
		if err != nil {
			//TODO: discuss if best effort is the right strategy here
			r.logger.Error().Str("op", "SyncAllRoutes").Msg(err.Error())
		}
	}
	return nil
}
