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
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	testLog "github.com/xmidt-org/ears/test/log"
	"net/http"
	"net/http/httptest"
	"testing"
)

type Validator struct {
	validate func(w http.ResponseWriter, r *http.Request)
}

func (v *Validator) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	v.validate(w, r)
}

func TestInitRequestMiddleware(t *testing.T) {

	listener := testLog.NewLogListener()
	logger := zerolog.New(listener)
	middleware := NewMiddleware(&logger)

	//Test Case 1
	validator := &Validator{func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		log.Ctx(ctx).Debug().Msg("test")
		listener.AssertLastLogLine(t, "tx.traceId", "123456")
	}}

	//3rd middleware should be the InitRequestMiddleware
	m := middleware[2](validator)

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.Header.Set("X-B3-TraceId", "123456")

	m.ServeHTTP(w, r)

	//Test Case 2
	validator = &Validator{func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		log.Ctx(ctx).Debug().Msg("test")
	}}

	//3rd middleware should be the InitRequestMiddleware
	m = middleware[2](validator)
	r = httptest.NewRequest(http.MethodGet, "/", nil)

	m.ServeHTTP(w, r)
}

func TestAuthMiddleware(t *testing.T) {
	ctx := context.Background()
	listener := testLog.NewLogListener()
	logger := zerolog.New(listener)
	middleware := NewMiddleware(&logger)
	subCtx := logger.WithContext(ctx)

	//AuthenticateMiddleware is currently just a pass through.
	//Just validate that it actually goes through
	validator := &Validator{func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		log.Ctx(ctx).Debug().Msg("good")
		listener.AssertLastLogLine(t, "message", "good")
	}}

	m := middleware[0](validator)

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.Header.Set("X-B3-TraceId", "123456")

	m.ServeHTTP(w, r.WithContext(subCtx))
}
