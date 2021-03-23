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
	"encoding/json"
	"errors"
	"github.com/xmidt-org/ears/internal/pkg/tablemgr"
	"io/ioutil"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/rs/zerolog/log"

	"github.com/xmidt-org/ears/pkg/route"
)

type APIManager struct {
	muxRouter       *mux.Router
	routingTableMgr tablemgr.RoutingTableManager
}

func NewAPIManager(routingMgr tablemgr.RoutingTableManager) (*APIManager, error) {
	api := &APIManager{
		muxRouter:       mux.NewRouter(),
		routingTableMgr: routingMgr,
	}
	api.muxRouter.HandleFunc("/ears/version", api.versionHandler).Methods(http.MethodGet)
	api.muxRouter.HandleFunc("/ears/v1/routes/{routeId}", api.addRouteHandler).Methods(http.MethodPut)
	api.muxRouter.HandleFunc("/ears/v1/routes", api.addRouteHandler).Methods(http.MethodPost)
	api.muxRouter.HandleFunc("/ears/v1/routes/{routeId}", api.removeRouteHandler).Methods(http.MethodDelete)
	api.muxRouter.HandleFunc("/ears/v1/routes/{routeId}", api.getRouteHandler).Methods(http.MethodGet)
	api.muxRouter.HandleFunc("/ears/v1/routes", api.getAllRoutesHandler).Methods(http.MethodGet)
	api.muxRouter.HandleFunc("/ears/v1/senders", api.getAllSendersHandler).Methods(http.MethodGet)
	api.muxRouter.HandleFunc("/ears/v1/receivers", api.getAllReceiversHandler).Methods(http.MethodGet)
	api.muxRouter.HandleFunc("/ears/v1/filters", api.getAllFiltersHandler).Methods(http.MethodGet)
	return api, nil
}

func (a *APIManager) versionHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log.Ctx(ctx).Debug().Msg("versionHandler")
	resp := ItemResponse(Version)
	resp.Respond(ctx, w)
}

func (a *APIManager) addRouteHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	routeId := vars["routeId"]
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Ctx(ctx).Error().Str("op", "addRouteHandler").Msg(err.Error())
		resp := ErrorResponse(err)
		resp.Respond(ctx, w)
		return
	}
	var route route.Config
	err = json.Unmarshal(body, &route)
	if err != nil {
		log.Ctx(ctx).Error().Str("op", "addRouteHandler").Msg(err.Error())
		resp := ErrorResponse(err)
		resp.Respond(ctx, w)
		return
	}
	if routeId != "" && route.Id != "" && routeId != route.Id {
		err := errors.New("route ID mismatch " + routeId + " vs " + route.Id)
		log.Ctx(ctx).Error().Str("op", "addRouteHandler").Msg(err.Error())
		resp := ErrorResponse(err)
		resp.Respond(ctx, w)
		return
	}
	if routeId != "" && route.Id == "" {
		route.Id = routeId
	}
	err = a.routingTableMgr.AddRoute(ctx, &route)
	if err != nil {
		log.Ctx(ctx).Error().Str("op", "addRouteHandler").Msg(err.Error())
		resp := ErrorResponse(err)
		resp.Respond(ctx, w)
		return
	}
	resp := ItemResponse(route)
	resp.Respond(ctx, w)
}

func (a *APIManager) removeRouteHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	routeId := vars["routeId"]
	err := a.routingTableMgr.RemoveRoute(ctx, routeId)
	if err != nil {
		log.Ctx(ctx).Error().Str("op", "removeRouteHandler").Msg(err.Error())
		resp := ErrorResponse(err)
		resp.Respond(ctx, w)
		return
	}
	resp := ItemResponse(routeId)
	resp.Respond(ctx, w)
}

func (a *APIManager) getRouteHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	routeId := vars["routeId"]
	routeConfig, err := a.routingTableMgr.GetRoute(ctx, routeId)
	if err != nil {
		log.Ctx(ctx).Error().Str("op", "getRouteHandler").Msg(err.Error())
		if _, ok := err.(*route.RouteNotFoundError); ok {
			resp := ErrorResponse(&NotFoundError{})
			resp.Respond(ctx, w)
		} else {
			resp := ErrorResponse(err)
			resp.Respond(ctx, w)
		}
		return

	}
	resp := ItemResponse(routeConfig)
	resp.Respond(ctx, w)
}

func (a *APIManager) getAllRoutesHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	allRouteConfigs, err := a.routingTableMgr.GetAllRoutes(ctx)
	if err != nil {
		log.Ctx(ctx).Error().Str("op", "getAllRoutesHandler").Msg(err.Error())
		resp := ErrorResponse(err)
		resp.Respond(ctx, w)
		return
	}
	resp := ItemsResponse(allRouteConfigs)
	resp.Respond(ctx, w)
}

func (a *APIManager) getAllSendersHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	allSenders, err := a.routingTableMgr.GetAllSenders(ctx)
	if err != nil {
		log.Ctx(ctx).Error().Str("op", "getAllSendersHandler").Msg(err.Error())
		resp := ErrorResponse(err)
		resp.Respond(ctx, w)
		return
	}
	resp := ItemsResponse(allSenders)
	resp.Respond(ctx, w)
}

func (a *APIManager) getAllReceiversHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	allReceivers, err := a.routingTableMgr.GetAllReceivers(ctx)
	if err != nil {
		log.Ctx(ctx).Error().Str("op", "getAllReceiversHandler").Msg(err.Error())
		resp := ErrorResponse(err)
		resp.Respond(ctx, w)
		return
	}
	resp := ItemsResponse(allReceivers)
	resp.Respond(ctx, w)
}

func (a *APIManager) getAllFiltersHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	allFilters, err := a.routingTableMgr.GetAllFilters(ctx)
	if err != nil {
		log.Ctx(ctx).Error().Str("op", "getAllFiltersHandler").Msg(err.Error())
		resp := ErrorResponse(err)
		resp.Respond(ctx, w)
		return
	}
	resp := ItemsResponse(allFilters)
	resp.Respond(ctx, w)
}
