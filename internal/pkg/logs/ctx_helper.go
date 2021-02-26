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

package logs

import (
	"context"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

//Helper function to derive context from a parent context in zerolog
func SubLoggerCtx(ctx context.Context, parent *zerolog.Logger) context.Context {
	subCtx := parent.WithContext(ctx)
	logger := log.Ctx(subCtx).With().Logger()
	return logger.WithContext(subCtx)
}

//Helper function to add zerolog key/value pari to a golang context
func StrToLogCtx(ctx context.Context, key string, value string) {
	l := zerolog.Ctx(ctx)
	l.UpdateContext(func(c zerolog.Context) zerolog.Context {
		return c.Str(key, value)
	})
}
