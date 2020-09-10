/**
 *  Copyright (c) 2020  Comcast Cable Communications Management, LLC
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package app

import (
	"context"
	"github.com/gorilla/mux"
	"github.com/rs/zerolog"
	"go.uber.org/fx"
	"net/http"
)

var Version = "v0.0.0"

func NewMux(r *mux.Router, middlewares []func(next http.Handler) http.Handler) (http.Handler, error) {
	for _, middleware := range middlewares {
		r.Use(middleware)
	}
	return r, nil
}

func SetupAPIServer(lifecycle fx.Lifecycle, config Config, logger *zerolog.Logger, mux http.Handler) {

	server := &http.Server{
		Addr:    ":" + config.GetString("api.port"),
		Handler: mux,
	}

	lifecycle.Append(
		fx.Hook{
			OnStart: func(context.Context) error {
				go server.ListenAndServe()
				logger.Info().Msg("API Server Started")
				return nil
			},
			OnStop: func(ctx context.Context) error {
				logger.Info().Msg("API Server Stopped")
				server.Shutdown(ctx)
				return nil
			},
		},
	)
}
