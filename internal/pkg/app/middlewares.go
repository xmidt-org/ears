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
	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"net/http"
)

var middlewareLogger *zerolog.Logger

func NewMiddleware(logger *zerolog.Logger) []func(next http.Handler) http.Handler {
	middlewareLogger = logger
	return []func(next http.Handler) http.Handler{
		initRequestMiddleware,
		authenticateMiddleware,
	}
}

func getSubCtxWithStr(ctx context.Context, key string, value string) context.Context {
	logger := log.Ctx(ctx).With().Str(key, value).Logger()
	return logger.WithContext(ctx)
}

func initRequestMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		subCtx := middlewareLogger.WithContext(ctx)

		traceId := r.Header.Get(HeaderTraceId)
		if traceId == "" {
			traceId = uuid.New().String()
		}
		subCtx = getSubCtxWithStr(subCtx, LogTraceId, traceId)

		appId := r.Header.Get(HeaderTenantId)
		if appId != "" {
			subCtx = getSubCtxWithStr(subCtx, LogTenantId, appId)
		}
		log.Ctx(subCtx).Debug().Msg("initializeRequestMiddleware")

		next.ServeHTTP(w, r.WithContext(subCtx))
	})
}

func authenticateMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		subCtx := r.Context()
		log.Ctx(subCtx).Debug().Msg("authenticateMiddleare")
		next.ServeHTTP(w, r.WithContext(subCtx))
	})
}
