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
	"github.com/xmidt-org/ears/internal/pkg/jwt"
	"go.uber.org/fx"
)

var Module = fx.Options(
	fx.Provide(
		ProvideJWTManager,
	),
)

type JWTIn struct {
	fx.In
	Config config.Config
	Logger *zerolog.Logger
}

type JWTOut struct {
	fx.Out
	JWTManager jwt.JWTConsumer
}

func ProvideJWTManager(in JWTIn) (JWTOut, error) {
	out := JWTOut{}
	requireBearerToken := in.Config.GetBool("ears.jwt.requireBearerToken")
	publicKeyEndpoint := in.Config.GetString("ears.jwt.publicKeyEndpoint")
	out.JWTManager, _ = jwt.NewJWTConsumer(publicKeyEndpoint, DefaultJWTVerifier, requireBearerToken)
	return out, nil
}

func DefaultJWTVerifier(path, method, scope string) bool {
	return true
}
