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
	"go.uber.org/fx"
	"sync"
	"time"
)

type DefaultRoutingTableManager struct {
	sync.Mutex
	pluginMgr    plugin.Manager
	storageMgr   route.RouteStorer
	rtSyncer     RoutingTableSyncer
	liveRouteMap map[string]*LiveRouteWrapper // in-memory references to live routes
	routeHashMap map[string]bool              // keep around the hashes of route configs
	logger       *zerolog.Logger
	config       Config
}

type LiveRouteWrapper struct {
	Route       *route.Route
	Sender      sender.Sender
	Receiver    receiver.Receiver
	FilterChain *filter.Chain
	Config      *route.Config
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

func SetupRoutingManager(lifecycle fx.Lifecycle, config Config, logger *zerolog.Logger, routingTableMgr RoutingTableManager) error {
	lifecycle.Append(
		fx.Hook{
			OnStart: func(context.Context) error {
				routingTableMgr.RegisterAllRoutes()
				routingTableMgr.StartListeningForSyncRequests()
				routingTableMgr.StartGlobalSyncChecker()
				logger.Info().Msg("Routing Manager Service Started")
				return nil
			},
			OnStop: func(ctx context.Context) error {
				routingTableMgr.StopListeningForSyncRequests()
				routingTableMgr.UnregisterAllRoutes()
				logger.Info().Msg("Routing Manager Stopped")
				return nil
			},
		},
	)
	return nil
}

func NewRoutingTableManager(pluginMgr plugin.Manager, storageMgr route.RouteStorer, logger *zerolog.Logger, config Config) RoutingTableManager {
	rtm := &DefaultRoutingTableManager{pluginMgr: pluginMgr, storageMgr: storageMgr, liveRouteMap: make(map[string]*LiveRouteWrapper, 0), routeHashMap: make(map[string]bool, 0), logger: logger, config: config}
	rtm.rtSyncer = NewRedisTableSyncer(rtm, logger, config)
	return rtm
}

func (r *DefaultRoutingTableManager) StartListeningForSyncRequests() {
	r.rtSyncer.StartListeningForSyncRequests()
}

func (r *DefaultRoutingTableManager) StopListeningForSyncRequests() {
	r.rtSyncer.StopListeningForSyncRequests()
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
	lrw.Config = routeConfig
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
	if routeId == "" {
		return errors.New("missing route ID")
	}
	routeConfig, err := r.storageMgr.GetRoute(ctx, routeId)
	// route does not even exist in persistence layer
	if err != nil {
		r.logger.Info().Str("op", "RemoveRoute").Str("routeId", routeConfig.Id).Msg("route already removed")
		return nil
	}
	routeHash := routeConfig.Hash(ctx)
	r.Lock()
	_, exists := r.routeHashMap[routeHash]
	delete(r.routeHashMap, routeHash)
	r.Unlock()
	// route does not exist as live route on this ears instance
	if !exists {
		r.logger.Info().Str("op", "RemoveRoute").Str("routeId", routeConfig.Id).Str("routeHash", routeHash).Msg("route already removed")
		return nil
	}
	err = r.storageMgr.DeleteRoute(ctx, routeId)
	if err != nil {
		return err
	}
	r.rtSyncer.PublishSyncRequest(ctx, routeId, false)
	return r.unregisterAndStopRoute(ctx, routeId)
}

func (r *DefaultRoutingTableManager) AddRoute(ctx context.Context, routeConfig *route.Config) error {
	ctx = context.Background()
	if routeConfig == nil {
		return errors.New("missing route config")
	}
	// use hashed ID if none is provided - this ID will be returned by the AddRoute REST API
	routeHash := routeConfig.Hash(ctx)
	if routeConfig.Id == "" {
		routeConfig.Id = routeHash
	}
	err := routeConfig.Validate(ctx)
	if err != nil {
		return err
	}
	// check if identical route (with same hash) already exists
	r.Lock()
	_, ok := r.routeHashMap[routeHash]
	r.routeHashMap[routeHash] = true
	r.Unlock()
	if ok {
		r.logger.Info().Str("op", "AddRoute").Str("routeId", routeConfig.Id).Str("routeHash", routeHash).Msg("route already exists")
		return nil
	}
	// currently storage layer handles created and updated timestamps
	err = r.storageMgr.SetRoute(ctx, *routeConfig)
	if err != nil {
		return err
	}
	r.rtSyncer.PublishSyncRequest(ctx, routeConfig.Id, true)
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

func (r *DefaultRoutingTableManager) StartGlobalSyncChecker() {
	go func() {
		time.Sleep(5 * time.Second)
		for {
			isSynchronized, err := r.IsSynchronized()
			if err == nil {
				if isSynchronized {
					r.logger.Info().Msg("routing table is synchronzied")
				} else {
					r.logger.Error().Msg("routing table is not synchronzied")
				}
			}
			time.Sleep(60 * time.Second)
		}
	}()
}

func (r *DefaultRoutingTableManager) IsSynchronized() (bool, error) {
	//TODO: should this be locked in any way?
	ctx := context.Background()
	storedRoutes, err := r.storageMgr.GetAllRoutes(ctx)
	if err != nil {
		return true, err
	}
	if len(storedRoutes) != len(r.routeHashMap) {
		return false, nil
	}
	for _, sr := range storedRoutes {
		_, ok := r.routeHashMap[sr.Hash(ctx)]
		if !ok {
			return false, nil
		}
	}
	return true, nil
}

func (r *DefaultRoutingTableManager) UnregisterAllRoutes() error {
	ctx := context.Background()
	r.logger.Info().Msg("starting to unregister all routes")
	var err error
	for _, lrw := range r.liveRouteMap {
		r.logger.Info().Msg("unregistering route " + lrw.Config.Id)
		r.Lock()
		delete(r.routeHashMap, lrw.Config.Hash(ctx))
		r.Unlock()
		err = r.unregisterAndStopRoute(ctx, lrw.Config.Id)
		if err != nil {
			//TODO: discuss if best effort is the right strategy here
			r.logger.Error().Str("op", "SyncAllRoutes").Msg(err.Error())
		}
	}
	r.logger.Info().Msg("done unregistering all routes")
	return nil
}

func (r *DefaultRoutingTableManager) RegisterAllRoutes() error {
	ctx := context.Background()
	r.logger.Info().Msg("starting to register all routes")
	var err error
	routeConfigs, err := r.storageMgr.GetAllRoutes(ctx)
	if err != nil {
		return err
	}
	//TODO: should this be locked in any way?
	for _, routeConfig := range routeConfigs {
		r.logger.Info().Msg("registering route " + routeConfig.Id)
		r.Lock()
		r.routeHashMap[routeConfig.Hash(ctx)] = true
		r.Unlock()
		err = r.registerAndRunRoute(ctx, &routeConfig)
		if err != nil {
			//TODO: discuss if best effort is the right strategy here
			r.logger.Error().Str("op", "SyncAllRoutes").Msg(err.Error())
		}
	}
	r.logger.Info().Msg("done registering all routes")
	return nil
}
