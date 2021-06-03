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
	"github.com/goccy/go-yaml"
	"github.com/xmidt-org/ears/internal/pkg/logs"
	"github.com/xmidt-org/ears/internal/pkg/quota"
	"github.com/xmidt-org/ears/internal/pkg/rtsemconv"
	"github.com/xmidt-org/ears/internal/pkg/tablemgr"
	"github.com/xmidt-org/ears/pkg/tenant"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/semconv"
	"go.opentelemetry.io/otel/trace"
	"io/ioutil"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/rs/zerolog/log"

	"github.com/xmidt-org/ears/pkg/route"
)

type APIManager struct {
	muxRouter       *mux.Router
	routingTableMgr tablemgr.RoutingTableManager
	tenantStorer    tenant.TenantStorer
	quotaManager    *quota.QuotaManager
}

func NewAPIManager(routingMgr tablemgr.RoutingTableManager, tenantStorer tenant.TenantStorer, quotaManager *quota.QuotaManager) (*APIManager, error) {
	api := &APIManager{
		muxRouter:       mux.NewRouter(),
		routingTableMgr: routingMgr,
		tenantStorer:    tenantStorer,
		quotaManager:    quotaManager,
	}
	api.muxRouter.HandleFunc("/ears/version", api.versionHandler).Methods(http.MethodGet)
	api.muxRouter.HandleFunc("/ears/v1/orgs/{orgId}/applications/{appId}/routes/{routeId}", api.addRouteHandler).Methods(http.MethodPut)
	api.muxRouter.HandleFunc("/ears/v1/orgs/{orgId}/applications/{appId}/routes", api.addRouteHandler).Methods(http.MethodPost)
	api.muxRouter.HandleFunc("/ears/v1/orgs/{orgId}/applications/{appId}/routes/{routeId}", api.removeRouteHandler).Methods(http.MethodDelete)
	api.muxRouter.HandleFunc("/ears/v1/orgs/{orgId}/applications/{appId}/routes/{routeId}", api.getRouteHandler).Methods(http.MethodGet)
	api.muxRouter.HandleFunc("/ears/v1/orgs/{orgId}/applications/{appId}/routes", api.getAllTenantRoutesHandler).Methods(http.MethodGet)
	api.muxRouter.HandleFunc("/ears/v1/orgs/{orgId}/applications/{appId}/config", api.getTenantConfigHandler).Methods(http.MethodGet)
	api.muxRouter.HandleFunc("/ears/v1/orgs/{orgId}/applications/{appId}/config", api.setTenantConfigHandler).Methods(http.MethodPut)
	api.muxRouter.HandleFunc("/ears/v1/orgs/{orgId}/applications/{appId}/config", api.deleteTenantConfigHandler).Methods(http.MethodDelete)
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

func getTenant(ctx context.Context, vars map[string]string) (*tenant.Id, ApiError) {
	orgId := vars["orgId"]
	appId := vars["appId"]
	logs.StrToLogCtx(ctx, "orgId", orgId)
	logs.StrToLogCtx(ctx, "appId", appId)
	if orgId == "" || appId == "" {
		var err ApiError
		if orgId == "" {
			err = &BadRequestError{"orgId empty", nil}
		} else {
			err = &BadRequestError{"appId empty", nil}
		}
		return nil, err
	}
	return &tenant.Id{OrgId: orgId, AppId: appId}, nil
}

func (a *APIManager) addRouteHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tracer := otel.Tracer("ears")
	ctx, span := tracer.Start(ctx, "addRoute")
	defer span.End()
	span.SetAttributes(semconv.HTTPMethodKey.String(r.Method))
	span.SetAttributes(semconv.HTTPTargetKey.String(r.RequestURI))
	span.SetAttributes(semconv.HTTPHostKey.String(r.Host))
	span.SetAttributes(semconv.HTTPRequestContentLengthKey.Int(int(r.ContentLength)))
	vars := mux.Vars(r)
	tid, apiErr := getTenant(ctx, vars)
	if apiErr != nil {
		log.Ctx(ctx).Error().Str("op", "AddRouteHandler").Str("error", apiErr.Error()).Msg("orgId or appId empty")
		resp := ErrorResponse(apiErr)
		resp.Respond(ctx, w)
		return
	}
	span.SetAttributes(rtsemconv.EARSOrgId.String(tid.OrgId))
	span.SetAttributes(rtsemconv.EARSAppId.String(tid.AppId))
	routeId := vars["routeId"]
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Ctx(ctx).Error().Str("op", "addRouteHandler").Msg(err.Error())
		resp := ErrorResponse(&InternalServerError{err})
		resp.Respond(ctx, w)
		return
	}
	var route route.Config
	err = yaml.Unmarshal(body, &route)
	if err != nil {
		log.Ctx(ctx).Error().Str("op", "addRouteHandler").Msg(err.Error())
		resp := ErrorResponse(&BadRequestError{"Cannot unmarshal request body", err})
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
	span.SetAttributes(rtsemconv.EARSRouteId.String(routeId))
	route.TenantId.AppId = tid.AppId
	route.TenantId.OrgId = tid.OrgId
	err = a.routingTableMgr.AddRoute(ctx, &route)
	if err != nil {
		log.Ctx(ctx).Error().Str("op", "addRouteHandler").Msg(err.Error())
		resp := ErrorResponse(convertToApiError(ctx, err))
		resp.Respond(ctx, w)
		return
	}
	span.SetAttributes(semconv.HTTPStatusCodeKey.Int(200))
	resp := ItemResponse(route)
	resp.Respond(ctx, w)
}

func (a *APIManager) removeRouteHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tracer := otel.Tracer("ears")
	ctx, span := tracer.Start(ctx, "removeRoute")
	defer span.End()
	span.SetAttributes(semconv.HTTPMethodKey.String(r.Method))
	span.SetAttributes(semconv.HTTPTargetKey.String(r.RequestURI))
	span.SetAttributes(semconv.HTTPHostKey.String(r.Host))
	vars := mux.Vars(r)
	tid, apiErr := getTenant(ctx, vars)
	if apiErr != nil {
		log.Ctx(ctx).Error().Str("op", "removeRouteHandler").Str("error", apiErr.Error()).Msg("orgId or appId empty")
		resp := ErrorResponse(apiErr)
		resp.Respond(ctx, w)
		return
	}
	span.SetAttributes(rtsemconv.EARSOrgId.String(tid.OrgId))
	span.SetAttributes(rtsemconv.EARSAppId.String(tid.AppId))
	routeId := vars["routeId"]
	span.SetAttributes(rtsemconv.EARSRouteId.String(routeId))
	err := a.routingTableMgr.RemoveRoute(ctx, *tid, routeId)
	if err != nil {
		log.Ctx(ctx).Error().Str("op", "removeRouteHandler").Msg(err.Error())
		resp := ErrorResponse(convertToApiError(ctx, err))
		resp.Respond(ctx, w)
		return
	}
	span.SetAttributes(semconv.HTTPStatusCodeKey.Int(200))
	resp := ItemResponse(routeId)
	resp.Respond(ctx, w)
}

func (a *APIManager) getRouteHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tracer := otel.Tracer("ears")
	ctx, span := tracer.Start(ctx, "getRoute")
	defer span.End()
	span.SetAttributes(semconv.HTTPMethodKey.String(r.Method))
	span.SetAttributes(semconv.HTTPTargetKey.String(r.RequestURI))
	span.SetAttributes(semconv.HTTPHostKey.String(r.Host))
	vars := mux.Vars(r)
	tid, apiErr := getTenant(ctx, vars)
	if apiErr != nil {
		log.Ctx(ctx).Error().Str("op", "getRouteHandler").Str("error", apiErr.Error()).Msg("orgId or appId empty")
		resp := ErrorResponse(apiErr)
		resp.Respond(ctx, w)
		return
	}
	span.SetAttributes(rtsemconv.EARSOrgId.String(tid.OrgId))
	span.SetAttributes(rtsemconv.EARSAppId.String(tid.AppId))
	routeId := vars["routeId"]
	span.SetAttributes(rtsemconv.EARSRouteId.String(routeId))
	routeConfig, err := a.routingTableMgr.GetRoute(ctx, *tid, routeId)
	if err != nil {
		log.Ctx(ctx).Error().Str("op", "getRouteHandler").Msg(err.Error())
		resp := ErrorResponse(convertToApiError(ctx, err))
		resp.Respond(ctx, w)
		return
	}
	span.SetAttributes(semconv.HTTPStatusCodeKey.Int(200))
	resp := ItemResponse(routeConfig)
	resp.Respond(ctx, w)
}

func (a *APIManager) getAllTenantRoutesHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tracer := otel.Tracer("ears")
	ctx, span := tracer.Start(ctx, "getAllTenantRoutes")
	defer span.End()
	span.SetAttributes(semconv.HTTPMethodKey.String(r.Method))
	span.SetAttributes(semconv.HTTPTargetKey.String(r.RequestURI))
	span.SetAttributes(semconv.HTTPHostKey.String(r.Host))
	vars := mux.Vars(r)
	tid, apiErr := getTenant(ctx, vars)
	if apiErr != nil {
		log.Ctx(ctx).Error().Str("op", "GetAllTenantRoutes").Str("error", apiErr.Error()).Msg("orgId or appId empty")
		resp := ErrorResponse(apiErr)
		resp.Respond(ctx, w)
		return
	}
	span.SetAttributes(rtsemconv.EARSOrgId.String(tid.OrgId))
	span.SetAttributes(rtsemconv.EARSAppId.String(tid.AppId))
	allRouteConfigs, err := a.routingTableMgr.GetAllTenantRoutes(ctx, *tid)
	if err != nil {
		log.Ctx(ctx).Error().Str("op", "GetAllTenantRoutes").Msg(err.Error())
		resp := ErrorResponse(convertToApiError(ctx, err))
		resp.Respond(ctx, w)
		return
	}
	span.SetAttributes(attribute.Int("routeCount", len(allRouteConfigs)))
	span.SetAttributes(semconv.HTTPStatusCodeKey.Int(200))
	resp := ItemsResponse(allRouteConfigs)
	resp.Respond(ctx, w)
}

func (a *APIManager) getAllSendersHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tracer := otel.Tracer("ears")
	ctx, span := tracer.Start(ctx, "getAllSenders")
	defer span.End()
	span.SetAttributes(semconv.HTTPMethodKey.String(r.Method))
	span.SetAttributes(semconv.HTTPTargetKey.String(r.RequestURI))
	span.SetAttributes(semconv.HTTPHostKey.String(r.Host))
	allSenders, err := a.routingTableMgr.GetAllSendersStatus(ctx)
	if err != nil {
		log.Ctx(ctx).Error().Str("op", "getAllSendersHandler").Msg(err.Error())
		resp := ErrorResponse(convertToApiError(ctx, err))
		resp.Respond(ctx, w)
		return
	}
	span.SetAttributes(attribute.Int("senderCount", len(allSenders)))
	span.SetAttributes(semconv.HTTPStatusCodeKey.Int(200))
	resp := ItemsResponse(allSenders)
	resp.Respond(ctx, w)
}

func (a *APIManager) getAllReceiversHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tracer := otel.Tracer("ears")
	ctx, span := tracer.Start(ctx, "getAllReceivers")
	defer span.End()
	span.SetAttributes(semconv.HTTPMethodKey.String(r.Method))
	span.SetAttributes(semconv.HTTPTargetKey.String(r.RequestURI))
	span.SetAttributes(semconv.HTTPHostKey.String(r.Host))
	allReceivers, err := a.routingTableMgr.GetAllReceiversStatus(ctx)
	if err != nil {
		log.Ctx(ctx).Error().Str("op", "getAllReceiversHandler").Msg(err.Error())
		resp := ErrorResponse(convertToApiError(ctx, err))
		resp.Respond(ctx, w)
		return
	}
	span.SetAttributes(attribute.Int("receiverCount", len(allReceivers)))
	span.SetAttributes(semconv.HTTPStatusCodeKey.Int(200))
	resp := ItemsResponse(allReceivers)
	resp.Respond(ctx, w)
}

func (a *APIManager) getAllFiltersHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tracer := otel.Tracer("ears")
	ctx, span := tracer.Start(ctx, "getAllFilters")
	defer span.End()
	span.SetAttributes(semconv.HTTPMethodKey.String(r.Method))
	span.SetAttributes(semconv.HTTPTargetKey.String(r.RequestURI))
	span.SetAttributes(semconv.HTTPHostKey.String(r.Host))
	allFilters, err := a.routingTableMgr.GetAllFiltersStatus(ctx)
	if err != nil {
		log.Ctx(ctx).Error().Str("op", "getAllFiltersHandler").Msg(err.Error())
		resp := ErrorResponse(convertToApiError(ctx, err))
		resp.Respond(ctx, w)
		return
	}
	span.SetAttributes(attribute.Int("filterCount", len(allFilters)))
	span.SetAttributes(semconv.HTTPStatusCodeKey.Int(200))
	resp := ItemsResponse(allFilters)
	resp.Respond(ctx, w)
}

func (a *APIManager) getTenantConfigHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tracer := otel.Tracer("ears")
	ctx, span := tracer.Start(ctx, "getTenant")
	defer span.End()
	span.SetAttributes(semconv.HTTPMethodKey.String(r.Method))
	span.SetAttributes(semconv.HTTPTargetKey.String(r.RequestURI))
	span.SetAttributes(semconv.HTTPHostKey.String(r.Host))
	vars := mux.Vars(r)
	tid, apiErr := getTenant(ctx, vars)
	if apiErr != nil {
		log.Ctx(ctx).Error().Str("op", "getTenantConfigHandler").Str("error", apiErr.Error()).Msg("orgId or appId empty")
		resp := ErrorResponse(apiErr)
		resp.Respond(ctx, w)
		return
	}
	span.SetAttributes(rtsemconv.EARSOrgId.String(tid.OrgId))
	span.SetAttributes(rtsemconv.EARSAppId.String(tid.AppId))
	config, err := a.tenantStorer.GetConfig(ctx, *tid)
	if err != nil {
		log.Ctx(ctx).Error().Str("op", "getTenantConfigHandler").Str("error", err.Error()).Msg("error getting tenant config")
		resp := ErrorResponse(convertToApiError(ctx, err))
		resp.Respond(ctx, w)
		return
	}
	span.SetAttributes(semconv.HTTPStatusCodeKey.Int(200))
	resp := ItemResponse(config)
	resp.Respond(ctx, w)
}

func (a *APIManager) setTenantConfigHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tracer := otel.Tracer("ears")
	ctx, span := tracer.Start(ctx, "addTenant")
	defer span.End()
	span.SetAttributes(semconv.HTTPMethodKey.String(r.Method))
	span.SetAttributes(semconv.HTTPTargetKey.String(r.RequestURI))
	span.SetAttributes(semconv.HTTPHostKey.String(r.Host))
	vars := mux.Vars(r)
	tid, apiErr := getTenant(ctx, vars)
	if apiErr != nil {
		log.Ctx(ctx).Error().Str("op", "setTenantConfigHandler").Str("error", apiErr.Error()).Msg("orgId or appId empty")
		resp := ErrorResponse(apiErr)
		resp.Respond(ctx, w)
		return
	}
	span.SetAttributes(rtsemconv.EARSOrgId.String(tid.OrgId))
	span.SetAttributes(rtsemconv.EARSAppId.String(tid.AppId))
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Ctx(ctx).Error().Str("op", "setTenantConfigHandler").Str("error", err.Error()).Msg("error reading request body")
		resp := ErrorResponse(&InternalServerError{err})
		resp.Respond(ctx, w)
		return
	}
	var tenantConfig tenant.Config
	err = yaml.Unmarshal(body, &tenantConfig)
	if err != nil {
		log.Ctx(ctx).Error().Str("op", "setTenantConfigHandler").Str("error", err.Error()).Msg("error unmarshal request body")
		err = &BadRequestError{"Cannot unmarshal request body", err}
		resp := ErrorResponse(&InternalServerError{err})
		resp.Respond(ctx, w)
		return
	}
	tenantConfig.Tenant = *tid
	err = a.tenantStorer.SetConfig(ctx, tenantConfig)
	if err != nil {
		log.Ctx(ctx).Error().Str("op", "setTenantConfigHandler").Str("error", err.Error()).Msg("error setting tenant config")
		resp := ErrorResponse(convertToApiError(ctx, err))
		resp.Respond(ctx, w)
		return
	}
	a.quotaManager.PublishQuota(ctx, *tid)
	span.SetAttributes(semconv.HTTPStatusCodeKey.Int(200))
	resp := ItemResponse(tenantConfig)
	resp.Respond(ctx, w)
}

func (a *APIManager) deleteTenantConfigHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tracer := otel.Tracer("ears")
	ctx, span := tracer.Start(ctx, "deleteTenant")
	defer span.End()
	span.SetAttributes(semconv.HTTPMethodKey.String(r.Method))
	span.SetAttributes(semconv.HTTPTargetKey.String(r.RequestURI))
	span.SetAttributes(semconv.HTTPHostKey.String(r.Host))
	vars := mux.Vars(r)
	tid, apiErr := getTenant(ctx, vars)
	if apiErr != nil {
		log.Ctx(ctx).Error().Str("op", "deleteTenantConfigHandler").Str("error", apiErr.Error()).Msg("orgId or appId empty")
		resp := ErrorResponse(apiErr)
		resp.Respond(ctx, w)
		return
	}
	span.SetAttributes(rtsemconv.EARSOrgId.String(tid.OrgId))
	span.SetAttributes(rtsemconv.EARSAppId.String(tid.AppId))
	err := a.tenantStorer.DeleteConfig(ctx, *tid)
	if err != nil {
		log.Ctx(ctx).Error().Str("op", "deleteTenantConfigHandler").Str("error", err.Error()).Msg("error deleting tenant config")
		resp := ErrorResponse(convertToApiError(ctx, err))
		resp.Respond(ctx, w)
		return
	}
	span.SetAttributes(semconv.HTTPStatusCodeKey.Int(200))
	resp := ItemResponse(tid)
	resp.Respond(ctx, w)
}

func convertToApiError(ctx context.Context, err error) ApiError {
	span := trace.SpanFromContext(ctx)
	var tenantNotFound *tenant.TenantNotFoundError
	var badTenantConfig *tenant.BadConfigError
	var badRouteConfig *tablemgr.BadConfigError
	var routeValidationError *tablemgr.RouteValidationError
	var routeRegistrationError *tablemgr.RouteRegistrationError
	var routeNotFound *route.RouteNotFoundError
	span.RecordError(err)
	if errors.As(err, &tenantNotFound) {
		span.SetAttributes(semconv.HTTPStatusCodeKey.Int(404))
		return &NotFoundError{"tenant " + tenantNotFound.Tenant.ToString() + " not found"}
	} else if errors.As(err, &badTenantConfig) {
		span.SetAttributes(semconv.HTTPStatusCodeKey.Int(400))
		return &BadRequestError{"bad tenant config", err}
	} else if errors.As(err, &badRouteConfig) {
		return &BadRequestError{"bad route config", err}
	} else if errors.As(err, &routeRegistrationError) {
		span.SetAttributes(semconv.HTTPStatusCodeKey.Int(400))
		return &BadRequestError{"bad route config", err}
	} else if errors.As(err, &routeValidationError) {
		span.SetAttributes(semconv.HTTPStatusCodeKey.Int(400))
		return &BadRequestError{"bad route config", err}
	} else if errors.As(err, &routeNotFound) {
		span.SetAttributes(semconv.HTTPStatusCodeKey.Int(404))
		return &NotFoundError{"route " + routeNotFound.RouteId + " not found"}
	}
	span.SetAttributes(semconv.HTTPStatusCodeKey.Int(500))
	return &InternalServerError{err}
}
