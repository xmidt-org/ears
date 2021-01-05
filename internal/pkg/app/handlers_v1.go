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
	"github.com/gorilla/mux"
	"github.com/rs/zerolog/log"
	"net/http"
)

type APIManager struct {
	muxRouter       *mux.Router
	routingTableMgr RoutingTableManager
}

func NewAPIManager(routingMgr RoutingTableManager) (*APIManager, error) {
	api := &APIManager{
		muxRouter:       mux.NewRouter(),
		routingTableMgr: routingMgr,
	}
	api.muxRouter.HandleFunc("/ears/version", api.versionHandler).Methods(http.MethodGet)
	api.muxRouter.HandleFunc("/ears/v1/routes/{route}", api.addRouteHandler).Methods(http.MethodPut)
	return api, nil
}

func (a *APIManager) versionHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	log.Ctx(ctx).Debug().Msg("versionHandler")
	resp := ItemResponse(Version)
	resp.Respond(ctx, w)
}

func (a *APIManager) addRouteHandler(w http.ResponseWriter, r *http.Request) {
	//TODO this handler is incomplete. It is here for demo purpose
	ctx := r.Context()
	err := a.routingTableMgr.AddRoute(ctx, nil)
	if err != nil {
		log.Ctx(ctx).Error().Str("op", "addRouteHandler").Msg(err.Error())
		resp := ErrorResponse(err)
		resp.Respond(ctx, w)
		return
	}
	resp := ErrorResponse(&NotImplementedError{})
	resp.Respond(ctx, w)
}
