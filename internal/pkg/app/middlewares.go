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
	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/xmidt-org/ears/pkg/panics"
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

func initRequestMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		subCtx := SubLoggerCtx(ctx, middlewareLogger)

		defer func() {
			p := recover()
			if p != nil {
				panicErr := panics.ToError(p)
				resp := ErrorResponse(panicErr)
				resp.Respond(subCtx, w)
				log.Ctx(subCtx).Error().Str("op", "initRequestMiddleware").Str("error", panicErr.Error()).
					Str("stackTrace", panicErr.StackTrace()).Msg("A panic has occurred")
			}
		}()

		traceId := r.Header.Get(HeaderTraceId)
		if traceId == "" {
			traceId = uuid.New().String()
		}
		StrToLogCtx(subCtx, LogTraceId, traceId)

		appId := r.Header.Get(HeaderTenantId)
		if appId != "" {
			StrToLogCtx(subCtx, LogTenantId, appId)
		}
		log.Ctx(subCtx).Debug().Msg("initializeRequestMiddleware")

		next.ServeHTTP(w, r.WithContext(subCtx))
	})
}

func authenticateMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		subCtx := r.Context()
		log.Ctx(subCtx).Debug().Msg("authenticateMiddleware")
		next.ServeHTTP(w, r.WithContext(subCtx))
	})
}
