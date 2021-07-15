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
	"github.com/xmidt-org/ears/internal/pkg/quota"
	"github.com/xmidt-org/ears/internal/pkg/rtsemconv"
	"github.com/xmidt-org/ears/internal/pkg/tablemgr"
	"github.com/xmidt-org/ears/pkg/app"
	logs2 "github.com/xmidt-org/ears/pkg/logs"
	"github.com/xmidt-org/ears/pkg/tenant"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/global"
	"go.opentelemetry.io/otel/trace"
	"io/ioutil"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/rs/zerolog/log"

	"github.com/xmidt-org/ears/pkg/route"
)

type APIManager struct {
	muxRouter                  *mux.Router
	routingTableMgr            tablemgr.RoutingTableManager
	tenantStorer               tenant.TenantStorer
	quotaManager               *quota.QuotaManager
	addRouteSuccessRecorder    metric.BoundFloat64Counter
	addRouteFailureRecorder    metric.BoundFloat64Counter
	removeRouteSuccessRecorder metric.BoundFloat64Counter
	removeRouteFailureRecorder metric.BoundFloat64Counter
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
	// metrics
	// where should meters live (api manager, uberfx, global variables,...)?
	meter := global.Meter(rtsemconv.EARSMeterName)
	// labels represent additional key-value descriptors that can be bound to a metric observer or recorder (huh?)
	commonLabels := []attribute.KeyValue{
		//attribute.String("labelFoo", "bar"),
	}
	// what about up/down counter?
	// metric recorders
	api.addRouteSuccessRecorder = metric.Must(meter).
		NewFloat64Counter(
			rtsemconv.EARSMetricAddRouteSuccess,
			metric.WithDescription("measures the number of routes added"),
		).Bind(commonLabels...)
	//defer addRouteSuccessRecorder.Unbind()
	api.addRouteFailureRecorder = metric.Must(meter).
		NewFloat64Counter(
			rtsemconv.EARSMetricAddRouteFailure,
			metric.WithDescription("measures the number of route add failures"),
		).Bind(commonLabels...)
	//defer addRouteFailureRecorder.Unbind()
	api.removeRouteSuccessRecorder = metric.Must(meter).
		NewFloat64Counter(
			rtsemconv.EARSMetricRemoveRouteSuccess,
			metric.WithDescription("measures the number of routes removed"),
		).Bind(commonLabels...)
	//defer removeRouteSuccessRecorder.Unbind()
	api.removeRouteFailureRecorder = metric.Must(meter).
		NewFloat64Counter(
			rtsemconv.EARSMetricRemoveRouteFailure,
			metric.WithDescription("measures the number of route remove failures"),
		).Bind(commonLabels...)
	//defer removeRouteFailureRecorder.Unbind()
	return api, nil
}

func (a *APIManager) versionHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log.Ctx(ctx).Debug().Msg("versionHandler")
	resp := ItemResponse(app.Version)
	resp.Respond(ctx, w)
}

func getTenant(ctx context.Context, vars map[string]string) (*tenant.Id, ApiError) {
	orgId := vars["orgId"]
	appId := vars["appId"]
	logs2.StrToLogCtx(ctx, "orgId", orgId)
	logs2.StrToLogCtx(ctx, "appId", appId)
	if orgId == "" || appId == "" {
		var err ApiError
		if orgId == "" {
			err = &BadRequestError{"orgId empty", nil}
		} else {
			err = &BadRequestError{"appId empty", nil}
		}
		return nil, err
	}
	span := trace.SpanFromContext(ctx)
	span.SetAttributes(rtsemconv.EARSOrgId.String(orgId))
	span.SetAttributes(rtsemconv.EARSAppId.String(appId))

	return &tenant.Id{OrgId: orgId, AppId: appId}, nil
}

func (a *APIManager) addRouteHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	//
	vars := mux.Vars(r)
	tid, apiErr := getTenant(ctx, vars)
	if apiErr != nil {
		log.Ctx(ctx).Error().Str("op", "AddRouteHandler").Str("error", apiErr.Error()).Msg("orgId or appId empty")
		a.addRouteFailureRecorder.Add(ctx, 1.0)
		resp := ErrorResponse(apiErr)
		resp.Respond(ctx, w)
		return
	}
	routeId := vars["routeId"]
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Ctx(ctx).Error().Str("op", "addRouteHandler").Msg(err.Error())
		a.addRouteFailureRecorder.Add(ctx, 1.0)
		resp := ErrorResponse(&InternalServerError{err})
		resp.Respond(ctx, w)
		return
	}
	var route route.Config
	err = yaml.Unmarshal(body, &route)
	if err != nil {
		log.Ctx(ctx).Error().Str("op", "addRouteHandler").Msg(err.Error())
		a.addRouteFailureRecorder.Add(ctx, 1.0)
		resp := ErrorResponse(&BadRequestError{"Cannot unmarshal request body", err})
		resp.Respond(ctx, w)
		return
	}
	if routeId != "" && route.Id != "" && routeId != route.Id {
		err := &BadRequestError{"route ID mismatch " + routeId + " vs " + route.Id, nil}
		log.Ctx(ctx).Error().Str("op", "addRouteHandler").Msg(err.Error())
		a.addRouteFailureRecorder.Add(ctx, 1.0)
		resp := ErrorResponse(err)
		resp.Respond(ctx, w)
		return
	}
	if routeId != "" && route.Id == "" {
		route.Id = routeId
	}
	trace.SpanFromContext(ctx).SetAttributes(rtsemconv.EARSRouteId.String(routeId))
	route.TenantId.AppId = tid.AppId
	route.TenantId.OrgId = tid.OrgId
	err = a.routingTableMgr.AddRoute(ctx, &route)
	if err != nil {
		log.Ctx(ctx).Error().Str("op", "addRouteHandler").Msg(err.Error())
		a.addRouteFailureRecorder.Add(ctx, 1.0)
		resp := ErrorResponse(convertToApiError(ctx, err))
		resp.Respond(ctx, w)
		return
	} else {
		a.addRouteSuccessRecorder.Add(ctx, 1.0)
	}
	resp := ItemResponse(route)
	resp.Respond(ctx, w)
}

func (a *APIManager) removeRouteHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	tid, apiErr := getTenant(ctx, vars)
	if apiErr != nil {
		log.Ctx(ctx).Error().Str("op", "removeRouteHandler").Str("error", apiErr.Error()).Msg("orgId or appId empty")
		a.removeRouteFailureRecorder.Add(ctx, 1.0)
		resp := ErrorResponse(apiErr)
		resp.Respond(ctx, w)
		return
	}
	routeId := vars["routeId"]
	trace.SpanFromContext(ctx).SetAttributes(rtsemconv.EARSRouteId.String(routeId))
	err := a.routingTableMgr.RemoveRoute(ctx, *tid, routeId)
	if err != nil {
		log.Ctx(ctx).Error().Str("op", "removeRouteHandler").Msg(err.Error())
		a.removeRouteFailureRecorder.Add(ctx, 1.0)
		resp := ErrorResponse(convertToApiError(ctx, err))
		resp.Respond(ctx, w)
		return
	} else {
		a.removeRouteSuccessRecorder.Add(ctx, 1.0)
	}
	resp := ItemResponse(routeId)
	resp.Respond(ctx, w)
}

func (a *APIManager) getRouteHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	tid, apiErr := getTenant(ctx, vars)
	if apiErr != nil {
		log.Ctx(ctx).Error().Str("op", "getRouteHandler").Str("error", apiErr.Error()).Msg("orgId or appId empty")
		resp := ErrorResponse(apiErr)
		resp.Respond(ctx, w)
		return
	}
	routeId := vars["routeId"]
	trace.SpanFromContext(ctx).SetAttributes(rtsemconv.EARSRouteId.String(routeId))
	routeConfig, err := a.routingTableMgr.GetRoute(ctx, *tid, routeId)
	if err != nil {
		log.Ctx(ctx).Error().Str("op", "getRouteHandler").Msg(err.Error())
		resp := ErrorResponse(convertToApiError(ctx, err))
		resp.Respond(ctx, w)
		return
	}
	resp := ItemResponse(routeConfig)
	resp.Respond(ctx, w)
}

func (a *APIManager) getAllTenantRoutesHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	tid, apiErr := getTenant(ctx, vars)
	if apiErr != nil {
		log.Ctx(ctx).Error().Str("op", "GetAllTenantRoutes").Str("error", apiErr.Error()).Msg("orgId or appId empty")
		resp := ErrorResponse(apiErr)
		resp.Respond(ctx, w)
		return
	}
	allRouteConfigs, err := a.routingTableMgr.GetAllTenantRoutes(ctx, *tid)
	if err != nil {
		log.Ctx(ctx).Error().Str("op", "GetAllTenantRoutes").Msg(err.Error())
		resp := ErrorResponse(convertToApiError(ctx, err))
		resp.Respond(ctx, w)
		return
	}
	trace.SpanFromContext(ctx).SetAttributes(attribute.Int("routeCount", len(allRouteConfigs)))
	resp := ItemsResponse(allRouteConfigs)
	resp.Respond(ctx, w)
}

func (a *APIManager) getAllSendersHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	allSenders, err := a.routingTableMgr.GetAllSendersStatus(ctx)
	if err != nil {
		log.Ctx(ctx).Error().Str("op", "getAllSendersHandler").Msg(err.Error())
		resp := ErrorResponse(convertToApiError(ctx, err))
		resp.Respond(ctx, w)
		return
	}
	trace.SpanFromContext(ctx).SetAttributes(attribute.Int("senderCount", len(allSenders)))
	resp := ItemsResponse(allSenders)
	resp.Respond(ctx, w)
}

func (a *APIManager) getAllReceiversHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	allReceivers, err := a.routingTableMgr.GetAllReceiversStatus(ctx)
	if err != nil {
		log.Ctx(ctx).Error().Str("op", "getAllReceiversHandler").Msg(err.Error())
		resp := ErrorResponse(convertToApiError(ctx, err))
		resp.Respond(ctx, w)
		return
	}
	trace.SpanFromContext(ctx).SetAttributes(attribute.Int("receiverCount", len(allReceivers)))
	resp := ItemsResponse(allReceivers)
	resp.Respond(ctx, w)
}

func (a *APIManager) getAllFiltersHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	allFilters, err := a.routingTableMgr.GetAllFiltersStatus(ctx)
	if err != nil {
		log.Ctx(ctx).Error().Str("op", "getAllFiltersHandler").Msg(err.Error())
		resp := ErrorResponse(convertToApiError(ctx, err))
		resp.Respond(ctx, w)
		return
	}
	trace.SpanFromContext(ctx).SetAttributes(attribute.Int("filterCount", len(allFilters)))
	resp := ItemsResponse(allFilters)
	resp.Respond(ctx, w)
}

func (a *APIManager) getTenantConfigHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	tid, apiErr := getTenant(ctx, vars)
	if apiErr != nil {
		log.Ctx(ctx).Error().Str("op", "getTenantConfigHandler").Str("error", apiErr.Error()).Msg("orgId or appId empty")
		resp := ErrorResponse(apiErr)
		resp.Respond(ctx, w)
		return
	}
	config, err := a.tenantStorer.GetConfig(ctx, *tid)
	if err != nil {
		log.Ctx(ctx).Error().Str("op", "getTenantConfigHandler").Str("error", err.Error()).Msg("error getting tenant config")
		resp := ErrorResponse(convertToApiError(ctx, err))
		resp.Respond(ctx, w)
		return
	}
	resp := ItemResponse(config)
	resp.Respond(ctx, w)
}

func (a *APIManager) setTenantConfigHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	vars := mux.Vars(r)
	tid, apiErr := getTenant(ctx, vars)
	if apiErr != nil {
		log.Ctx(ctx).Error().Str("op", "setTenantConfigHandler").Str("error", apiErr.Error()).Msg("orgId or appId empty")
		resp := ErrorResponse(apiErr)
		resp.Respond(ctx, w)
		return
	}

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
	resp := ItemResponse(tenantConfig)
	resp.Respond(ctx, w)
}

func (a *APIManager) deleteTenantConfigHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	tid, apiErr := getTenant(ctx, vars)
	if apiErr != nil {
		log.Ctx(ctx).Error().Str("op", "deleteTenantConfigHandler").Str("error", apiErr.Error()).Msg("orgId or appId empty")
		resp := ErrorResponse(apiErr)
		resp.Respond(ctx, w)
		return
	}
	err := a.tenantStorer.DeleteConfig(ctx, *tid)
	if err != nil {
		log.Ctx(ctx).Error().Str("op", "deleteTenantConfigHandler").Str("error", err.Error()).Msg("error deleting tenant config")
		resp := ErrorResponse(convertToApiError(ctx, err))
		resp.Respond(ctx, w)
		return
	}
	resp := ItemResponse(tid)
	resp.Respond(ctx, w)
}

func convertToApiError(ctx context.Context, err error) ApiError {
	span := trace.SpanFromContext(ctx)
	span.RecordError(err)

	var tenantNotFound *tenant.TenantNotFoundError
	var badTenantConfig *tenant.BadConfigError
	var badRouteConfig *tablemgr.BadConfigError
	var routeValidationError *tablemgr.RouteValidationError
	var routeRegistrationError *tablemgr.RouteRegistrationError
	var routeNotFound *route.RouteNotFoundError
	if errors.As(err, &tenantNotFound) {
		return &NotFoundError{"tenant " + tenantNotFound.Tenant.ToString() + " not found"}
	} else if errors.As(err, &badTenantConfig) {
		return &BadRequestError{"bad tenant config", err}
	} else if errors.As(err, &badRouteConfig) {
		return &BadRequestError{"bad route config", err}
	} else if errors.As(err, &routeRegistrationError) {
		return &BadRequestError{"bad route config", err}
	} else if errors.As(err, &routeValidationError) {
		return &BadRequestError{"bad route config", err}
	} else if errors.As(err, &routeNotFound) {
		return &NotFoundError{"route " + routeNotFound.RouteId + " not found"}
	}
	return &InternalServerError{err}
}
