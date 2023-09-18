// Copyright 2021 Comcast Cable Communications Management, LLC
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

package jwtmanagerfx

import (
	"github.com/rs/zerolog"
	"github.com/xmidt-org/ears/internal/pkg/config"
	jwt2 "github.com/xmidt-org/ears/pkg/jwt"
	"github.com/xmidt-org/ears/pkg/tenant"
	"go.uber.org/fx"
	"regexp"
	"strings"
)

var Module = fx.Options(
	fx.Provide(
		ProvideJWTManager,
	),
)

type JWTIn struct {
	fx.In
	Config       config.Config
	Logger       *zerolog.Logger
	TenantStorer tenant.TenantStorer
}

type JWTOut struct {
	fx.Out
	JWTManager jwt2.JWTConsumer
}

func ProvideJWTManager(in JWTIn) (JWTOut, error) {
	out := JWTOut{}
	requireBearerToken := in.Config.GetBool("ears.jwt.requireBearerToken")
	publicKeyEndpoint := in.Config.GetString("ears.jwt.publicKeyEndpoint")
	domain := in.Config.GetString("ears.jwt.domain")
	component := in.Config.GetString("ears.jwt.component")
	adminClientIds := []string{}
	if in.Config.GetString("ears.jwt.adminClientIds") != "" {
		adminClientIds = strings.Split(in.Config.GetString("ears.jwt.adminClientIds"), ",")
	}
	capabilityPrefixes := []string{}
	if in.Config.GetString("ears.jwt.capabilityPrefixes") != "" {
		capabilityPrefixes = strings.Split(in.Config.GetString("ears.jwt.capabilityPrefixes"), ",")
	}
	out.JWTManager, _ = jwt2.NewJWTConsumer(publicKeyEndpoint, DefaultJWTVerifier, requireBearerToken, domain, component, adminClientIds, capabilityPrefixes, in.TenantStorer)
	return out, nil
}

func DefaultJWTVerifier(path, method, scope string) bool {
	if scope == "*:*" {
		return true
	}
	if scope == "routes:*" {
		r1, _ := regexp.Compile(`([\w]+/)*routes`)
		r2, _ := regexp.Compile(`([\w]+/)*routes/[\w]+`)
		if r1.MatchString(path) || r2.MatchString(path) {
			return true
		}
	}
	return false
}
