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
	"github.com/gorilla/mux"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/xmidt-org/ears/internal/pkg/jwt"
	"github.com/xmidt-org/ears/internal/pkg/rtsemconv"
	"github.com/xmidt-org/ears/pkg/logs"
	"github.com/xmidt-org/ears/pkg/panics"
	"github.com/xmidt-org/ears/pkg/route"
	"github.com/xmidt-org/ears/pkg/tenant"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gorilla/mux/otelmux"
	"go.opentelemetry.io/contrib/propagators/b3"
	"net/http"
	"regexp"
	"strings"
)

var middlewareLogger *zerolog.Logger
var jwtMgr jwt.JWTConsumer
var eventUrlValidator = regexp.MustCompile(`^\/ears\/v1\/orgs\/` + tenant.ORG_ID_REGEX + `\/applications\/` + tenant.APP_ID_REGEX + `\/routes\/` + route.ROUTE_ID_REGEX + `\/event$`)

func NewMiddleware(logger *zerolog.Logger, jwtManager jwt.JWTConsumer) []func(next http.Handler) http.Handler {
	middlewareLogger = logger
	jwtMgr = jwtManager
	otelMiddleware := otelmux.Middleware("ears", otelmux.WithPropagators(b3.New()))

	return []func(next http.Handler) http.Handler{
		authenticateMiddleware,
		otelMiddleware,
		initRequestMiddleware,
	}
}

func initRequestMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		subCtx := logs.SubLoggerCtx(ctx, middlewareLogger)

		defer func() {
			p := recover()
			if p != nil {
				panicErr := panics.ToError(p)
				resp := ErrorResponse(&InternalServerError{panicErr})
				resp.Respond(subCtx, w, doYaml(r))
				log.Ctx(subCtx).Error().Str("op", "initRequestMiddleware").Str("error", panicErr.Error()).
					Str("stackTrace", panicErr.StackTrace()).Msg("A panic has occurred")
			}
		}()

		traceId := r.Header.Get(HeaderTraceId)
		if traceId == "" {
			traceId = uuid.New().String()
		}
		logs.StrToLogCtx(subCtx, rtsemconv.EarsLogTraceIdKey, traceId)

		log.Ctx(subCtx).Debug().Msg("initializeRequestMiddleware")

		next.ServeHTTP(w, r.WithContext(subCtx))
	})
}

func authenticateMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		if r.URL.Path == "/ears/version" {
			next.ServeHTTP(w, r)
			return
		}
		if strings.HasPrefix(r.URL.Path, "/ears/openapi") {
			next.ServeHTTP(w, r)
			return
		}
		var tid *tenant.Id
		if strings.HasPrefix(r.URL.Path, "/ears/v1/routes") ||
			strings.HasPrefix(r.URL.Path, "/ears/v1/senders") ||
			strings.HasPrefix(r.URL.Path, "/ears/v1/receivers") ||
			strings.HasPrefix(r.URL.Path, "/ears/v1/filters") ||
			strings.HasPrefix(r.URL.Path, "/ears/v1/fragments") ||
			strings.HasPrefix(r.URL.Path, "/ears/v1/tenants") {
		} else {
			var tenantErr ApiError
			vars := mux.Vars(r)
			tid, tenantErr = getTenant(ctx, vars)
			if tenantErr != nil {
				log.Ctx(ctx).Error().Str("op", "authenticateMiddleware").Str("error", tenantErr.Error()).Msg("orgId or appId empty")
				resp := ErrorResponse(tenantErr)
				resp.Respond(ctx, w, doYaml(r))
				return
			}
			// do not authenticate event API calls here
			if eventUrlValidator.MatchString(r.URL.Path) {
				next.ServeHTTP(w, r)
				return
			}
		}
		bearerToken := getBearerToken(r)
		_, _, authErr := jwtMgr.VerifyToken(ctx, bearerToken, r.URL.Path, r.Method, tid)
		if authErr != nil {
			log.Ctx(ctx).Error().Str("op", "authenticateMiddleware").Str("error", authErr.Error()).Msg("authorization error")
			resp := ErrorResponse(convertToApiError(ctx, authErr))
			resp.Respond(ctx, w, doYaml(r))
			return
		}
		next.ServeHTTP(w, r)
	})
}
