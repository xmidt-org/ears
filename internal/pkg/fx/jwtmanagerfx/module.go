// SPDX-FileCopyrightText: 2025 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package jwtmanagerfx

import (
	"regexp"
	"strings"

	"github.com/rs/zerolog"
	"github.com/xmidt-org/ears/internal/pkg/config"
	"github.com/xmidt-org/ears/pkg/jwt"
	"github.com/xmidt-org/ears/pkg/tenant"
	"go.uber.org/fx"
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
	JWTManager jwt.JWTConsumer
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
	out.JWTManager, _ = jwt.NewJWTConsumer(publicKeyEndpoint, DefaultJWTVerifier, requireBearerToken, domain, component, adminClientIds, capabilityPrefixes, in.TenantStorer)
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
