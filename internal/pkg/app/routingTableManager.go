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
	"fmt"
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
	liveRouteMap map[string]*LiveRouteWrapper // references to live routes by route ID
	routeHashMap map[string]*LiveRouteWrapper // references to live routes by hash
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
				routingTableMgr.StartListeningForSyncRequests()
				routingTableMgr.RegisterAllRoutes()
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
	rtm := &DefaultRoutingTableManager{pluginMgr: pluginMgr, storageMgr: storageMgr, liveRouteMap: make(map[string]*LiveRouteWrapper, 0), routeHashMap: make(map[string]*LiveRouteWrapper, 0), logger: logger, config: config}
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
		r.logger.Info().Str("op", "unregisterAndStopRoute").Str("routeId", routeId).Msg("no live route exists with this ID")
		// no error to make this idempotent
		//return errors.New("no live route exists with ID " + routeId)
	} else {
		r.Lock()
		delete(r.liveRouteMap, routeId)
		delete(r.routeHashMap, liveRoute.Config.Hash(ctx))
		r.Unlock()
		err = liveRoute.Route.Stop(ctx)
		if err != nil {
			r.logger.Error().Str("op", "unregisterAndStopRoute").Msg(err.Error())
			//return err
		}
		err = liveRoute.Unregister(ctx, r)
		if err != nil {
			return err
		}
		r.logger.Info().Str("op", "unregisterAndStopRoute").Str("routeId", routeId).Msg("route stopped")
	}
	return nil
}

func (r *DefaultRoutingTableManager) registerAndRunRoute(ctx context.Context, routeConfig *route.Config) error {
	var err error
	var lrw LiveRouteWrapper
	lrw.Config = routeConfig
	// check if route already exists, check if this is an update etc.
	existingLiveRoute, ok := r.liveRouteMap[routeConfig.Id]
	if ok && existingLiveRoute.Config.Hash(ctx) == routeConfig.Hash(ctx) {
		r.logger.Info().Str("op", "registerAndRunRoute").Str("routeId", routeConfig.Id).Msg("identical route exists with same ID")
		return nil
	}
	// it's an update
	if ok && existingLiveRoute.Config.Hash(ctx) != routeConfig.Hash(ctx) {
		r.logger.Info().Str("op", "registerAndRunRoute").Str("routeId", routeConfig.Id).Msg("existing route needs to be updated")
		err = r.unregisterAndStopRoute(ctx, routeConfig.Id)
		if err != nil {
			r.logger.Info().Str("op", "registerAndRunRoute").Str("routeId", routeConfig.Id).Msg(err.Error())
		}
	}
	// an identical route already exists under a different ID - what to do here?
	existingLiveRoute, ok = r.routeHashMap[routeConfig.Hash(ctx)]
	if ok {
		r.logger.Info().Str("op", "registerAndRunRoute").Str("routeId", routeConfig.Id).Msg("identical route already exists under different ID " + existingLiveRoute.Config.Id)
		return nil
	}
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
	r.routeHashMap[routeConfig.Hash(ctx)] = &lrw
	r.Unlock()
	r.logger.Info().Str("op", "registerAndRunRoute").Str("routeId", routeConfig.Id).Msg("starting route")
	go func() {
		sctx := context.Background()
		//TODO: use application context here? see issue #51
		err = lrw.Route.Run(sctx, lrw.Receiver, lrw.FilterChain, lrw.Sender) // run is blocking
		if err != nil {
			r.logger.Error().Str("op", "registerAndRunRoute").Msg(err.Error())
		}
	}()
	return nil
}

func (r *DefaultRoutingTableManager) RemoveRoute(ctx context.Context, routeId string) error {
	if routeId == "" {
		return errors.New("missing route ID")
	}
	err := r.storageMgr.DeleteRoute(ctx, routeId)
	if err != nil {
		r.logger.Info().Str("op", "RemoveRoute").Str("routeId", routeId).Msg("could not delete route from storage layer")
		//return err
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
			cnt, err := r.Synchronize()
			if err != nil {
				r.logger.Error().Str("op", "StartGlobalSyncChecker").Msg(err.Error())
			}
			if cnt > 0 {
				r.logger.Error().Str("op", "StartGlobalSyncChecker").Msg(fmt.Sprintf("inconsistency detecting, %d entries in routing table repaired", cnt))
			} else {
				r.logger.Info().Str("op", "StartGlobalSyncChecker").Msg("routing table is synchronized")
			}
			time.Sleep(60 * time.Second)
		}
	}()
}

func (r *DefaultRoutingTableManager) IsSynchronized() (bool, error) {
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

/*func (r *DefaultRoutingTableManager) Synchronize() (int, error) {
	ok, err := r.IsSynchronized()
	if err != nil {
		return 0, err
	}
	if !ok {
		r.UnregisterAllRoutes()
		r.RegisterAllRoutes()
		return 1, nil
	}
	return 0, nil
}*/

func (r *DefaultRoutingTableManager) Synchronize() (int, error) {
	ctx := context.Background()
	storedRoutes, err := r.storageMgr.GetAllRoutes(ctx)
	if err != nil {
		return 0, err
	}
	storedRouteMap := make(map[string]route.Config, 0)
	for _, storedRoute := range storedRoutes {
		storedRouteMap[storedRoute.Id] = storedRoute
	}
	mutated := 0
	// stop all inconsistent or deleted routes
	for _, liveRoute := range r.liveRouteMap {
		storedRoute, ok := storedRouteMap[liveRoute.Config.Id]
		if !ok {
			r.logger.Info().Str("op", "Synchronize").Str("routeId", liveRoute.Config.Id).Msg("route stopped")
			r.unregisterAndStopRoute(ctx, liveRoute.Config.Id)
			mutated++
		} else if liveRoute.Config.Hash(ctx) != storedRoute.Hash(ctx) {
			r.logger.Info().Str("op", "Synchronize").Str("routeId", liveRoute.Config.Id).Msg("route stopped")
			r.unregisterAndStopRoute(ctx, liveRoute.Config.Id)
			mutated++
		}
	}
	// start all missing routes
	for _, storedRoute := range storedRoutes {
		_, ok := r.liveRouteMap[storedRoute.Id]
		if !ok {
			r.logger.Info().Str("op", "Synchronize").Str("routeId", storedRoute.Id).Msg("route started")
			var rc route.Config
			rc = storedRoute
			r.registerAndRunRoute(ctx, &rc)
			mutated++
		}
	}
	return mutated, nil
}

func (r *DefaultRoutingTableManager) UnregisterAllRoutes() error {
	ctx := context.Background()
	r.logger.Info().Str("op", "UnregisterAllRoutes").Msg("starting to unregister all routes")
	var err error
	for _, lrw := range r.liveRouteMap {
		r.logger.Info().Str("op", "UnregisterAllRoutes").Msg("unregistering route " + lrw.Config.Id)
		err = r.unregisterAndStopRoute(ctx, lrw.Config.Id)
		if err != nil {
			// best effort strategy
			r.logger.Error().Str("op", "UnregisterAllRoutes").Msg(err.Error())
		}
	}
	r.logger.Info().Str("op", "UnregisterAllRoutes").Msg("done unregistering all routes")
	return nil
}

func (r *DefaultRoutingTableManager) RegisterAllRoutes() error {
	ctx := context.Background()
	r.logger.Info().Str("op", "RegisterAllRoutes").Msg("starting to register all routes")
	var err error
	routeConfigs, err := r.storageMgr.GetAllRoutes(ctx)
	if err != nil {
		return err
	}
	for _, routeConfig := range routeConfigs {
		var rc route.Config
		rc = routeConfig
		r.logger.Info().Str("op", "RegisterAllRoutes").Msg("registering route " + routeConfig.Id)
		err = r.registerAndRunRoute(ctx, &rc)
		if err != nil {
			// best effort strategy
			r.logger.Error().Str("op", "RegisterAllRoutes").Msg(err.Error())
		}
	}
	r.logger.Info().Str("op", "RegisterAllRoutes").Msg("done registering all routes")
	return nil
}
