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
	"embed"
	"errors"
	"github.com/goccy/go-yaml"
	"github.com/xmidt-org/ears/internal/pkg/jwt"
	"github.com/xmidt-org/ears/internal/pkg/plugin"
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
	"strings"

	"github.com/gorilla/mux"
	"github.com/rs/zerolog/log"

	"github.com/xmidt-org/ears/pkg/route"
)

//go:embed ears
var WebsiteFS embed.FS

type APIManager struct {
	muxRouter                  *mux.Router
	routingTableMgr            tablemgr.RoutingTableManager
	tenantStorer               tenant.TenantStorer
	quotaManager               *quota.QuotaManager
	jwtManager                 jwt.JWTConsumer
	addRouteSuccessRecorder    metric.BoundFloat64Counter
	addRouteFailureRecorder    metric.BoundFloat64Counter
	removeRouteSuccessRecorder metric.BoundFloat64Counter
	removeRouteFailureRecorder metric.BoundFloat64Counter
}

func NewAPIManager(routingMgr tablemgr.RoutingTableManager, tenantStorer tenant.TenantStorer, quotaManager *quota.QuotaManager, jwtManager jwt.JWTConsumer) (*APIManager, error) {
	api := &APIManager{
		muxRouter:       mux.NewRouter(),
		routingTableMgr: routingMgr,
		tenantStorer:    tenantStorer,
		quotaManager:    quotaManager,
		jwtManager:      jwtManager,
	}
	api.muxRouter.PathPrefix("/ears/openapi").Handler(
		http.FileServer(http.FS(WebsiteFS)),
	)
	api.muxRouter.HandleFunc("/ears/version", api.versionHandler).Methods(http.MethodGet)

	api.muxRouter.HandleFunc("/ears/v1/orgs/{orgId}/applications/{appId}/routes/{routeId}", api.addRouteHandler).Methods(http.MethodPut)
	api.muxRouter.HandleFunc("/ears/v1/orgs/{orgId}/applications/{appId}/routes", api.addRouteHandler).Methods(http.MethodPost)
	api.muxRouter.HandleFunc("/ears/v1/orgs/{orgId}/applications/{appId}/routes/{routeId}", api.removeRouteHandler).Methods(http.MethodDelete)
	api.muxRouter.HandleFunc("/ears/v1/orgs/{orgId}/applications/{appId}/routes/{routeId}", api.getRouteHandler).Methods(http.MethodGet)
	api.muxRouter.HandleFunc("/ears/v1/orgs/{orgId}/applications/{appId}/routes", api.getAllTenantRoutesHandler).Methods(http.MethodGet)

	api.muxRouter.HandleFunc("/ears/v1/orgs/{orgId}/applications/{appId}/senders", api.getAllSendersHandler).Methods(http.MethodGet)
	api.muxRouter.HandleFunc("/ears/v1/orgs/{orgId}/applications/{appId}/receivers", api.getAllReceiversHandler).Methods(http.MethodGet)
	api.muxRouter.HandleFunc("/ears/v1/orgs/{orgId}/applications/{appId}/filters", api.getAllFiltersHandler).Methods(http.MethodGet)

	api.muxRouter.HandleFunc("/ears/v1/orgs/{orgId}/applications/{appId}/fragments/{fragmentId}", api.addFragmentHandler).Methods(http.MethodPut)
	api.muxRouter.HandleFunc("/ears/v1/orgs/{orgId}/applications/{appId}/fragments", api.addFragmentHandler).Methods(http.MethodPost)
	api.muxRouter.HandleFunc("/ears/v1/orgs/{orgId}/applications/{appId}/fragments/{fragmentId}", api.removeFragmentHandler).Methods(http.MethodDelete)
	api.muxRouter.HandleFunc("/ears/v1/orgs/{orgId}/applications/{appId}/fragments/{fragmentId}", api.getFragmentHandler).Methods(http.MethodGet)
	api.muxRouter.HandleFunc("/ears/v1/orgs/{orgId}/applications/{appId}/fragments", api.getAllTenantFragmentsHandler).Methods(http.MethodGet)

	api.muxRouter.HandleFunc("/ears/v1/orgs/{orgId}/applications/{appId}/config", api.getTenantConfigHandler).Methods(http.MethodGet)
	api.muxRouter.HandleFunc("/ears/v1/orgs/{orgId}/applications/{appId}/config", api.setTenantConfigHandler).Methods(http.MethodPut)
	api.muxRouter.HandleFunc("/ears/v1/orgs/{orgId}/applications/{appId}/config", api.deleteTenantConfigHandler).Methods(http.MethodDelete)
	api.muxRouter.HandleFunc("/ears/v1/routes", api.getAllRoutesHandler).Methods(http.MethodGet)

	api.muxRouter.HandleFunc("/ears/v1/tenants", api.getAllTenantConfigsHandler).Methods(http.MethodGet)
	api.muxRouter.HandleFunc("/ears/v1/senders", api.getAllSendersHandler).Methods(http.MethodGet)
	api.muxRouter.HandleFunc("/ears/v1/receivers", api.getAllReceiversHandler).Methods(http.MethodGet)
	api.muxRouter.HandleFunc("/ears/v1/filters", api.getAllFiltersHandler).Methods(http.MethodGet)
	api.muxRouter.HandleFunc("/ears/v1/fragments", api.getAllFragmentsHandler).Methods(http.MethodGet)
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

func doYaml(r *http.Request) bool {
	ct := r.Header.Get("Content-Type")
	return strings.Contains(ct, "yaml")
}

func (a *APIManager) versionHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log.Ctx(ctx).Debug().Msg("versionHandler")
	resp := ItemResponse(app.Version)
	resp.Respond(ctx, w, doYaml(r))
}

func getBearerToken(req *http.Request) string {
	return strings.TrimPrefix(req.Header.Get("Authorization"), "Bearer ")
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
	vars := mux.Vars(r)
	tid, apiErr := getTenant(ctx, vars)
	if apiErr != nil {
		log.Ctx(ctx).Error().Str("op", "addRouteHandler").Str("error", apiErr.Error()).Msg("orgId or appId empty")
		a.addRouteFailureRecorder.Add(ctx, 1.0)
		resp := ErrorResponse(apiErr)
		resp.Respond(ctx, w, doYaml(r))
		return
	}
	bearerToken := getBearerToken(r)
	_, _, authErr := a.jwtManager.VerifyToken(ctx, bearerToken, r.URL.Path, r.Method, tid)
	if authErr != nil {
		log.Ctx(ctx).Error().Str("op", "addRouteHandler").Str("error", authErr.Error()).Msg("authorization error")
		resp := ErrorResponse(convertToApiError(ctx, authErr))
		resp.Respond(ctx, w, doYaml(r))
		return
	}
	_, err := a.tenantStorer.GetConfig(ctx, *tid)
	if err != nil {
		log.Ctx(ctx).Error().Str("op", "addRouteHandler").Str("error", err.Error()).Msg("error getting tenant config")
		resp := ErrorResponse(convertToApiError(ctx, err))
		resp.Respond(ctx, w, doYaml(r))
		return
	}
	routeId := vars["routeId"]
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Ctx(ctx).Error().Str("op", "addRouteHandler").Msg(err.Error())
		a.addRouteFailureRecorder.Add(ctx, 1.0)
		resp := ErrorResponse(&InternalServerError{err})
		resp.Respond(ctx, w, doYaml(r))
		return
	}
	var route route.Config
	err = yaml.Unmarshal(body, &route)
	if err != nil {
		log.Ctx(ctx).Error().Str("op", "addRouteHandler").Msg(err.Error())
		a.addRouteFailureRecorder.Add(ctx, 1.0)
		resp := ErrorResponse(&BadRequestError{"Cannot unmarshal request body", err})
		resp.Respond(ctx, w, doYaml(r))
		return
	}
	if routeId != "" && route.Id != "" && routeId != route.Id {
		err := &BadRequestError{"route ID mismatch " + routeId + " vs " + route.Id, nil}
		log.Ctx(ctx).Error().Str("op", "addRouteHandler").Msg(err.Error())
		a.addRouteFailureRecorder.Add(ctx, 1.0)
		resp := ErrorResponse(err)
		resp.Respond(ctx, w, doYaml(r))
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
		resp.Respond(ctx, w, doYaml(r))
		return
	} else {
		a.addRouteSuccessRecorder.Add(ctx, 1.0)
	}
	resp := ItemResponse(route)
	resp.Respond(ctx, w, doYaml(r))
}

func (a *APIManager) removeRouteHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	tid, apiErr := getTenant(ctx, vars)
	if apiErr != nil {
		log.Ctx(ctx).Error().Str("op", "removeRouteHandler").Str("error", apiErr.Error()).Msg("orgId or appId empty")
		a.removeRouteFailureRecorder.Add(ctx, 1.0)
		resp := ErrorResponse(apiErr)
		resp.Respond(ctx, w, doYaml(r))
		return
	}
	bearerToken := getBearerToken(r)
	_, _, authErr := a.jwtManager.VerifyToken(ctx, bearerToken, r.URL.Path, r.Method, tid)
	if authErr != nil {
		log.Ctx(ctx).Error().Str("op", "removeRouteHandler").Str("error", authErr.Error()).Msg("authorization error")
		resp := ErrorResponse(convertToApiError(ctx, authErr))
		resp.Respond(ctx, w, doYaml(r))
		return
	}
	routeId := vars["routeId"]
	trace.SpanFromContext(ctx).SetAttributes(rtsemconv.EARSRouteId.String(routeId))
	err := a.routingTableMgr.RemoveRoute(ctx, *tid, routeId)
	if err != nil {
		log.Ctx(ctx).Error().Str("op", "removeRouteHandler").Msg(err.Error())
		a.removeRouteFailureRecorder.Add(ctx, 1.0)
		resp := ErrorResponse(convertToApiError(ctx, err))
		resp.Respond(ctx, w, doYaml(r))
		return
	} else {
		a.removeRouteSuccessRecorder.Add(ctx, 1.0)
	}
	resp := ItemResponse(routeId)
	resp.Respond(ctx, w, doYaml(r))
}

func (a *APIManager) getRouteHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	tid, apiErr := getTenant(ctx, vars)
	if apiErr != nil {
		log.Ctx(ctx).Error().Str("op", "getRouteHandler").Str("error", apiErr.Error()).Msg("orgId or appId empty")
		resp := ErrorResponse(apiErr)
		resp.Respond(ctx, w, doYaml(r))
		return
	}
	routeId := vars["routeId"]
	trace.SpanFromContext(ctx).SetAttributes(rtsemconv.EARSRouteId.String(routeId))
	routeConfig, err := a.routingTableMgr.GetRoute(ctx, *tid, routeId)
	if err != nil {
		log.Ctx(ctx).Error().Str("op", "getRouteHandler").Msg(err.Error())
		resp := ErrorResponse(convertToApiError(ctx, err))
		resp.Respond(ctx, w, doYaml(r))
		return
	}
	resp := ItemResponse(routeConfig)
	resp.Respond(ctx, w, doYaml(r))
}

func (a *APIManager) getAllTenantRoutesHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	tid, apiErr := getTenant(ctx, vars)
	if apiErr != nil {
		log.Ctx(ctx).Error().Str("op", "GetAllTenantRoutes").Str("error", apiErr.Error()).Msg("orgId or appId empty")
		resp := ErrorResponse(apiErr)
		resp.Respond(ctx, w, doYaml(r))
		return
	}
	allRouteConfigs, err := a.routingTableMgr.GetAllTenantRoutes(ctx, *tid)
	if err != nil {
		log.Ctx(ctx).Error().Str("op", "GetAllTenantRoutes").Msg(err.Error())
		resp := ErrorResponse(convertToApiError(ctx, err))
		resp.Respond(ctx, w, doYaml(r))
		return
	}
	trace.SpanFromContext(ctx).SetAttributes(attribute.Int("routeCount", len(allRouteConfigs)))
	resp := ItemsResponse(allRouteConfigs)
	resp.Respond(ctx, w, doYaml(r))
}

func (a *APIManager) getAllRoutesHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	allRouteConfigs := make([]route.Config, 0)
	configs, err := a.tenantStorer.GetAllConfigs(ctx)
	if err != nil {
		log.Ctx(ctx).Error().Str("op", "GetAllRoutes").Str("error", err.Error()).Msg("tenant configs read error")
		resp := ErrorResponse(convertToApiError(ctx, err))
		resp.Respond(ctx, w, doYaml(r))
		return
	}
	for _, config := range configs {
		tenantRouteConfigs, err := a.routingTableMgr.GetAllTenantRoutes(ctx, config.Tenant)
		if err != nil {
			log.Ctx(ctx).Error().Str("op", "GetAllRoutes").Msg(err.Error())
			resp := ErrorResponse(convertToApiError(ctx, err))
			resp.Respond(ctx, w, doYaml(r))
			return
		}
		allRouteConfigs = append(allRouteConfigs, tenantRouteConfigs...)
	}
	trace.SpanFromContext(ctx).SetAttributes(attribute.Int("routeCount", len(allRouteConfigs)))
	resp := ItemsResponse(allRouteConfigs)
	resp.Respond(ctx, w, doYaml(r))
}

func (a *APIManager) getAllTenantFragmentsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	tid, apiErr := getTenant(ctx, vars)
	if apiErr != nil {
		log.Ctx(ctx).Error().Str("op", "GetAllTenantFragments").Str("error", apiErr.Error()).Msg("orgId or appId empty")
		resp := ErrorResponse(apiErr)
		resp.Respond(ctx, w, doYaml(r))
		return
	}
	allFragments, err := a.routingTableMgr.GetAllTenantFragments(ctx, *tid)
	if err != nil {
		log.Ctx(ctx).Error().Str("op", "GetAllTenantFragments").Msg(err.Error())
		resp := ErrorResponse(convertToApiError(ctx, err))
		resp.Respond(ctx, w, doYaml(r))
		return
	}
	trace.SpanFromContext(ctx).SetAttributes(attribute.Int("routeCount", len(allFragments)))
	resp := ItemsResponse(allFragments)
	resp.Respond(ctx, w, doYaml(r))
}

func (a *APIManager) getAllSendersHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	allSenders, err := a.routingTableMgr.GetAllSendersStatus(ctx)
	if err != nil {
		log.Ctx(ctx).Error().Str("op", "getAllSendersHandler").Msg(err.Error())
		resp := ErrorResponse(convertToApiError(ctx, err))
		resp.Respond(ctx, w, doYaml(r))
		return
	}
	senders := make(map[string]plugin.SenderStatus)
	tid, _ := getTenant(ctx, vars)
	if tid != nil {
		for k, v := range allSenders {
			if tid.Equal(v.Tid) {
				senders[k] = v
			}
		}
	} else {
		senders = allSenders
	}
	trace.SpanFromContext(ctx).SetAttributes(attribute.Int("senderCount", len(senders)))
	resp := ItemsResponse(senders)
	resp.Respond(ctx, w, doYaml(r))
}

func (a *APIManager) getAllReceiversHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	allReceivers, err := a.routingTableMgr.GetAllReceiversStatus(ctx)
	if err != nil {
		log.Ctx(ctx).Error().Str("op", "getAllReceiversHandler").Msg(err.Error())
		resp := ErrorResponse(convertToApiError(ctx, err))
		resp.Respond(ctx, w, doYaml(r))
		return
	}
	receivers := make(map[string]plugin.ReceiverStatus)
	tid, _ := getTenant(ctx, vars)
	if tid != nil {
		for k, v := range allReceivers {
			if tid.Equal(v.Tid) {
				receivers[k] = v
			}
		}
	} else {
		receivers = allReceivers
	}
	trace.SpanFromContext(ctx).SetAttributes(attribute.Int("receiverCount", len(receivers)))
	resp := ItemsResponse(receivers)
	resp.Respond(ctx, w, doYaml(r))
}

func (a *APIManager) getAllFiltersHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	allFilters, err := a.routingTableMgr.GetAllFiltersStatus(ctx)
	if err != nil {
		log.Ctx(ctx).Error().Str("op", "getAllFiltersHandler").Msg(err.Error())
		resp := ErrorResponse(convertToApiError(ctx, err))
		resp.Respond(ctx, w, doYaml(r))
		return
	}
	filters := make(map[string]plugin.FilterStatus)
	tid, _ := getTenant(ctx, vars)
	if tid != nil {
		for k, v := range allFilters {
			if tid.Equal(v.Tid) {
				filters[k] = v
			}
		}
	} else {
		filters = allFilters
	}
	trace.SpanFromContext(ctx).SetAttributes(attribute.Int("filterCount", len(filters)))
	resp := ItemsResponse(filters)
	resp.Respond(ctx, w, doYaml(r))
}

func (a *APIManager) getAllFragmentsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	allFragments, err := a.routingTableMgr.GetAllFragments(ctx)
	if err != nil {
		log.Ctx(ctx).Error().Str("op", "getAllFragmentsHandler").Msg(err.Error())
		resp := ErrorResponse(convertToApiError(ctx, err))
		resp.Respond(ctx, w, doYaml(r))
		return
	}
	trace.SpanFromContext(ctx).SetAttributes(attribute.Int("fragmentCount", len(allFragments)))
	resp := ItemsResponse(allFragments)
	resp.Respond(ctx, w, doYaml(r))
}

func (a *APIManager) getFragmentHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	tid, apiErr := getTenant(ctx, vars)
	if apiErr != nil {
		log.Ctx(ctx).Error().Str("op", "getFragmentHandler").Str("error", apiErr.Error()).Msg("orgId or appId empty")
		resp := ErrorResponse(apiErr)
		resp.Respond(ctx, w, doYaml(r))
		return
	}
	fragmentId := vars["fragmentId"]
	trace.SpanFromContext(ctx).SetAttributes(rtsemconv.EARSFragmentId.String(fragmentId))
	fragmentConfig, err := a.routingTableMgr.GetFragment(ctx, *tid, fragmentId)
	if err != nil {
		log.Ctx(ctx).Error().Str("op", "getFragmentHandler").Msg(err.Error())
		resp := ErrorResponse(convertToApiError(ctx, err))
		resp.Respond(ctx, w, doYaml(r))
		return
	}
	resp := ItemResponse(fragmentConfig)
	resp.Respond(ctx, w, doYaml(r))
}

func (a *APIManager) removeFragmentHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	tid, apiErr := getTenant(ctx, vars)
	if apiErr != nil {
		log.Ctx(ctx).Error().Str("op", "removeFragmentHandler").Str("error", apiErr.Error()).Msg("orgId or appId empty")
		a.removeRouteFailureRecorder.Add(ctx, 1.0)
		resp := ErrorResponse(apiErr)
		resp.Respond(ctx, w, doYaml(r))
		return
	}
	bearerToken := getBearerToken(r)
	_, _, authErr := a.jwtManager.VerifyToken(ctx, bearerToken, r.URL.Path, r.Method, tid)
	if authErr != nil {
		log.Ctx(ctx).Error().Str("op", "removeFragmentHandler").Str("error", authErr.Error()).Msg("authorization error")
		resp := ErrorResponse(convertToApiError(ctx, authErr))
		resp.Respond(ctx, w, doYaml(r))
		return
	}
	fragmentId := vars["fragmentId"]
	trace.SpanFromContext(ctx).SetAttributes(rtsemconv.EARSFragmentId.String(fragmentId))
	err := a.routingTableMgr.RemoveFragment(ctx, *tid, fragmentId)
	if err != nil {
		log.Ctx(ctx).Error().Str("op", "removeFragmentHandler").Msg(err.Error())
		a.removeRouteFailureRecorder.Add(ctx, 1.0)
		resp := ErrorResponse(convertToApiError(ctx, err))
		resp.Respond(ctx, w, doYaml(r))
		return
	} else {
		a.removeRouteSuccessRecorder.Add(ctx, 1.0)
	}
	resp := ItemResponse(fragmentId)
	resp.Respond(ctx, w, doYaml(r))
}

func (a *APIManager) addFragmentHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	tid, apiErr := getTenant(ctx, vars)
	if apiErr != nil {
		log.Ctx(ctx).Error().Str("op", "addFragmentHandler").Str("error", apiErr.Error()).Msg("orgId or appId empty")
		a.addRouteFailureRecorder.Add(ctx, 1.0)
		resp := ErrorResponse(apiErr)
		resp.Respond(ctx, w, doYaml(r))
		return
	}
	bearerToken := getBearerToken(r)
	_, _, authErr := a.jwtManager.VerifyToken(ctx, bearerToken, r.URL.Path, r.Method, tid)
	if authErr != nil {
		log.Ctx(ctx).Error().Str("op", "addFragmentHandler").Str("error", authErr.Error()).Msg("authorization error")
		resp := ErrorResponse(convertToApiError(ctx, authErr))
		resp.Respond(ctx, w, doYaml(r))
		return
	}
	_, err := a.tenantStorer.GetConfig(ctx, *tid)
	if err != nil {
		log.Ctx(ctx).Error().Str("op", "addFragmentHandler").Str("error", err.Error()).Msg("error getting tenant config")
		resp := ErrorResponse(convertToApiError(ctx, err))
		resp.Respond(ctx, w, doYaml(r))
		return
	}
	fragmentId := vars["fragmentId"]
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Ctx(ctx).Error().Str("op", "addFragmentHandler").Msg(err.Error())
		a.addRouteFailureRecorder.Add(ctx, 1.0)
		resp := ErrorResponse(&InternalServerError{err})
		resp.Respond(ctx, w, doYaml(r))
		return
	}
	var fragmentConfig route.PluginConfig
	err = yaml.Unmarshal(body, &fragmentConfig)
	if err != nil {
		log.Ctx(ctx).Error().Str("op", "addFragmentHandler").Msg(err.Error())
		resp := ErrorResponse(&BadRequestError{"Cannot unmarshal request body", err})
		resp.Respond(ctx, w, doYaml(r))
		return
	}
	if fragmentId != "" && fragmentConfig.FragmentName != "" && fragmentId != fragmentConfig.FragmentName {
		err := &BadRequestError{"fragment name mismatch " + fragmentId + " vs " + fragmentConfig.Name, nil}
		log.Ctx(ctx).Error().Str("op", "addFragmentHandler").Msg(err.Error())
		resp := ErrorResponse(err)
		resp.Respond(ctx, w, doYaml(r))
		return
	}
	if fragmentId != "" && fragmentConfig.FragmentName == "" {
		fragmentConfig.FragmentName = fragmentId
	}
	if fragmentConfig.FragmentName == "" {
		err := &BadRequestError{"missing fragment name", nil}
		log.Ctx(ctx).Error().Str("op", "addFragmentHandler").Msg(err.Error())
		resp := ErrorResponse(err)
		resp.Respond(ctx, w, doYaml(r))
		return
	}
	trace.SpanFromContext(ctx).SetAttributes(rtsemconv.EARSFragmentId.String(fragmentId))
	err = a.routingTableMgr.AddFragment(ctx, *tid, fragmentConfig)
	if err != nil {
		log.Ctx(ctx).Error().Str("op", "addFragmentHandler").Msg(err.Error())
		resp := ErrorResponse(convertToApiError(ctx, err))
		resp.Respond(ctx, w, doYaml(r))
		return
	}
	resp := ItemResponse(fragmentConfig)
	resp.Respond(ctx, w, doYaml(r))
}

func (a *APIManager) getTenantConfigHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	tid, apiErr := getTenant(ctx, vars)
	if apiErr != nil {
		log.Ctx(ctx).Error().Str("op", "getTenantConfigHandler").Str("error", apiErr.Error()).Msg("orgId or appId empty")
		resp := ErrorResponse(apiErr)
		resp.Respond(ctx, w, doYaml(r))
		return
	}
	config, err := a.tenantStorer.GetConfig(ctx, *tid)
	if err != nil {
		log.Ctx(ctx).Error().Str("op", "getTenantConfigHandler").Str("error", err.Error()).Msg("error getting tenant config")
		resp := ErrorResponse(convertToApiError(ctx, err))
		resp.Respond(ctx, w, doYaml(r))
		return
	}
	resp := ItemResponse(config)
	resp.Respond(ctx, w, doYaml(r))
}

func (a *APIManager) getAllTenantConfigsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	configs, err := a.tenantStorer.GetAllConfigs(ctx)
	if err != nil {
		log.Ctx(ctx).Error().Str("op", "getAllTenantConfigsHandler").Str("error", err.Error()).Msg("error getting all tenant configs")
		resp := ErrorResponse(convertToApiError(ctx, err))
		resp.Respond(ctx, w, doYaml(r))
		return
	}
	resp := ItemsResponse(configs)
	resp.Respond(ctx, w, doYaml(r))
}

func (a *APIManager) setTenantConfigHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	tid, apiErr := getTenant(ctx, vars)
	if apiErr != nil {
		log.Ctx(ctx).Error().Str("op", "setTenantConfigHandler").Str("error", apiErr.Error()).Msg("orgId or appId empty")
		resp := ErrorResponse(apiErr)
		resp.Respond(ctx, w, doYaml(r))
		return
	}
	bearerToken := getBearerToken(r)
	_, _, authErr := a.jwtManager.VerifyToken(ctx, bearerToken, r.URL.Path, r.Method, tid)
	if authErr != nil {
		log.Ctx(ctx).Error().Str("op", "setTenantConfigHandler").Str("error", authErr.Error()).Msg("authorization error")
		resp := ErrorResponse(convertToApiError(ctx, authErr))
		resp.Respond(ctx, w, doYaml(r))
		return
	}
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Ctx(ctx).Error().Str("op", "setTenantConfigHandler").Str("error", err.Error()).Msg("error reading request body")
		resp := ErrorResponse(&InternalServerError{err})
		resp.Respond(ctx, w, doYaml(r))
		return
	}
	var tenantConfig tenant.Config
	err = yaml.Unmarshal(body, &tenantConfig)
	if err != nil {
		log.Ctx(ctx).Error().Str("op", "setTenantConfigHandler").Str("error", err.Error()).Msg("error unmarshal request body")
		err = &BadRequestError{"Cannot unmarshal request body", err}
		resp := ErrorResponse(&InternalServerError{err})
		resp.Respond(ctx, w, doYaml(r))
		return
	}
	tenantConfig.Tenant = *tid
	err = a.tenantStorer.SetConfig(ctx, tenantConfig)
	if err != nil {
		log.Ctx(ctx).Error().Str("op", "setTenantConfigHandler").Str("error", err.Error()).Msg("error setting tenant config")
		resp := ErrorResponse(convertToApiError(ctx, err))
		resp.Respond(ctx, w, doYaml(r))
		return
	}
	a.quotaManager.PublishQuota(ctx, *tid)
	resp := ItemResponse(tenantConfig)
	resp.Respond(ctx, w, doYaml(r))
}

func (a *APIManager) deleteTenantConfigHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	tid, apiErr := getTenant(ctx, vars)
	if apiErr != nil {
		log.Ctx(ctx).Error().Str("op", "deleteTenantConfigHandler").Str("error", apiErr.Error()).Msg("orgId or appId empty")
		resp := ErrorResponse(apiErr)
		resp.Respond(ctx, w, doYaml(r))
		return
	}
	bearerToken := getBearerToken(r)
	_, _, authErr := a.jwtManager.VerifyToken(ctx, bearerToken, r.URL.Path, r.Method, tid)
	if authErr != nil {
		log.Ctx(ctx).Error().Str("op", "deleteTenantConfigHandler").Str("error", authErr.Error()).Msg("authorization error")
		resp := ErrorResponse(convertToApiError(ctx, authErr))
		resp.Respond(ctx, w, doYaml(r))
		return
	}
	allRouteConfigs, err := a.routingTableMgr.GetAllTenantRoutes(ctx, *tid)
	if err != nil {
		log.Ctx(ctx).Error().Str("op", "deleteTenantConfigHandler").Msg(err.Error())
		resp := ErrorResponse(convertToApiError(ctx, err))
		resp.Respond(ctx, w, doYaml(r))
		return
	}
	if len(allRouteConfigs) > 0 {
		log.Ctx(ctx).Error().Str("op", "deleteTenantConfigHandler").Msg("tenant has routes")
		resp := ErrorResponse(convertToApiError(ctx, &BadRequestError{"tenant has routes", nil}))
		resp.Respond(ctx, w, doYaml(r))
		return
	}
	err = a.tenantStorer.DeleteConfig(ctx, *tid)
	if err != nil {
		log.Ctx(ctx).Error().Str("op", "deleteTenantConfigHandler").Str("error", err.Error()).Msg("error deleting tenant config")
		resp := ErrorResponse(convertToApiError(ctx, err))
		resp.Respond(ctx, w, doYaml(r))
		return
	}
	resp := ItemResponse(tid)
	resp.Respond(ctx, w, doYaml(r))
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
	var jwtAuthError *jwt.JWTAuthError
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
	} else if errors.As(err, &jwtAuthError) {
		return &BadRequestError{"bad or missing jwt token", err}
	}
	return &InternalServerError{err}
}
