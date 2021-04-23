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
	"github.com/goccy/go-yaml"
	"github.com/xmidt-org/ears/internal/pkg/logs"
	"github.com/xmidt-org/ears/internal/pkg/tablemgr"
	"github.com/xmidt-org/ears/pkg/tenant"
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
	api.muxRouter.HandleFunc("/ears/v1/orgs/{orgId}/applications/{appId}/routes/{routeId}", api.addRouteHandler).Methods(http.MethodPut)
	api.muxRouter.HandleFunc("/ears/v1/orgs/{orgId}/applications/{appId}/routes", api.addRouteHandler).Methods(http.MethodPost)
	api.muxRouter.HandleFunc("/ears/v1/orgs/{orgId}/applications/{appId}/routes/{routeId}", api.removeRouteHandler).Methods(http.MethodDelete)
	api.muxRouter.HandleFunc("/ears/v1/orgs/{orgId}/applications/{appId}/routes/{routeId}", api.getRouteHandler).Methods(http.MethodGet)
	api.muxRouter.HandleFunc("/ears/v1/orgs/{orgId}/applications/{appId}/routes", api.getAllTenantRoutesHandler).Methods(http.MethodGet)
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

func getTenant(ctx context.Context, vars map[string]string) (*tenant.Id, error) {
	orgId := vars["orgId"]
	appId := vars["appId"]
	logs.StrToLogCtx(ctx, "orgId", orgId)
	logs.StrToLogCtx(ctx, "appId", appId)

	if orgId == "" || appId == "" {
		var err error
		if orgId == "" {
			err = &BadRequestError{"orgId empty", nil}
		} else {
			err = &BadRequestError{"appId empty", nil}
		}
		return nil, err
	}
	return &tenant.Id{orgId, appId}, nil
}

func (a *APIManager) addRouteHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)

	tid, err := getTenant(ctx, vars)
	if err != nil {
		log.Ctx(ctx).Error().Str("op", "AddRouteHandler").Str("error", err.Error()).Msg("orgId or appId empty")
		resp := ErrorResponse(err)
		resp.Respond(ctx, w)
		return
	}
	routeId := vars["routeId"]
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Ctx(ctx).Error().Str("op", "addRouteHandler").Msg(err.Error())
		resp := ErrorResponse(&BadRequestError{err.Error(), err})
		resp.Respond(ctx, w)
		return
	}
	var route route.Config
	//err = json.Unmarshal(body, &route)
	err = yaml.Unmarshal(body, &route)
	if err != nil {
		log.Ctx(ctx).Error().Str("op", "addRouteHandler").Msg(err.Error())
		err = &BadRequestError{"Cannot unmarshal request body", err}
		resp := ErrorResponse(&BadRequestError{err.Error(), err})
		resp.Respond(ctx, w)
		return
	}
	if routeId != "" && route.Id != "" && routeId != route.Id {
		err := &BadRequestError{"route ID mismatch " + routeId + " vs " + route.Id, nil}
		log.Ctx(ctx).Error().Str("op", "addRouteHandler").Msg(err.Error())
		resp := ErrorResponse(err)
		resp.Respond(ctx, w)
		return
	}
	if routeId != "" && route.Id == "" {
		route.Id = routeId
	}
	route.TenantId.AppId = tid.AppId
	route.TenantId.OrgId = tid.OrgId
	err = a.routingTableMgr.AddRoute(ctx, &route)
	if err != nil {
		log.Ctx(ctx).Error().Str("op", "addRouteHandler").Msg(err.Error())
		resp := ErrorResponse(&BadRequestError{err.Error(), err})
		resp.Respond(ctx, w)
		return
	}
	resp := ItemResponse(route)
	resp.Respond(ctx, w)
}

func (a *APIManager) removeRouteHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	tid, err := getTenant(ctx, vars)
	if err != nil {
		log.Ctx(ctx).Error().Str("op", "removeRouteHandler").Str("error", err.Error()).Msg("orgId or appId empty")
		resp := ErrorResponse(err)
		resp.Respond(ctx, w)
		return
	}
	routeId := vars["routeId"]
	err = a.routingTableMgr.RemoveRoute(ctx, *tid, routeId)
	if err != nil {
		log.Ctx(ctx).Error().Str("op", "removeRouteHandler").Msg(err.Error())
		resp := ErrorResponse(&BadRequestError{err.Error(), err})
		resp.Respond(ctx, w)
		return
	}
	resp := ItemResponse(routeId)
	resp.Respond(ctx, w)
}

func (a *APIManager) getRouteHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	tid, err := getTenant(ctx, vars)
	if err != nil {
		log.Ctx(ctx).Error().Str("op", "getRouteHandler").Str("error", err.Error()).Msg("orgId or appId empty")
		resp := ErrorResponse(err)
		resp.Respond(ctx, w)
		return
	}
	routeId := vars["routeId"]
	routeConfig, err := a.routingTableMgr.GetRoute(ctx, *tid, routeId)
	if err != nil {
		log.Ctx(ctx).Error().Str("op", "getRouteHandler").Msg(err.Error())
		if _, ok := err.(*route.RouteNotFoundError); ok {
			resp := ErrorResponse(&NotFoundError{})
			resp.Respond(ctx, w)
		} else {
			resp := ErrorResponse(&BadRequestError{err.Error(), err})
			resp.Respond(ctx, w)
		}
		return

	}
	resp := ItemResponse(routeConfig)
	resp.Respond(ctx, w)
}

func (a *APIManager) getAllTenantRoutesHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	tid, err := getTenant(ctx, vars)
	if err != nil {
		log.Ctx(ctx).Error().Str("op", "GetAllTenantRoutes").Str("error", err.Error()).Msg("orgId or appId empty")
		resp := ErrorResponse(err)
		resp.Respond(ctx, w)
		return
	}
	allRouteConfigs, err := a.routingTableMgr.GetAllTenantRoutes(ctx, *tid)
	if err != nil {
		log.Ctx(ctx).Error().Str("op", "GetAllTenantRoutes").Msg(err.Error())
		resp := ErrorResponse(&BadRequestError{err.Error(), err})
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
		resp := ErrorResponse(&BadRequestError{err.Error(), err})
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
		resp := ErrorResponse(&BadRequestError{err.Error(), err})
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
		resp := ErrorResponse(&BadRequestError{err.Error(), err})
		resp.Respond(ctx, w)
		return
	}
	resp := ItemsResponse(allFilters)
	resp.Respond(ctx, w)
}
