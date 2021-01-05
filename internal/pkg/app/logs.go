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
	"fmt"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"os"
	"sync"
)

var appLogger *zerolog.Logger
var appLoggerLock = &sync.Mutex{}

func ProvideLogger(config Config) (*zerolog.Logger, error) {
	appLoggerLock.Lock()
	defer appLoggerLock.Unlock()

	if appLogger != nil {
		return appLogger, nil
	}

	logLevel, err := zerolog.ParseLevel(config.GetString("ears.logLevel"))
	if err != nil {
		return nil, &InvalidOptionError{
			Option: fmt.Sprintf("loglevel %s is not valid", config.GetString("ears.logLevel")),
		}
	}
	logger := zerolog.New(os.Stdout).Level(logLevel)
	zerolog.LevelFieldName = "log.level"
	appLogger = &logger
	return appLogger, nil
}

func SubLoggerCtx(ctx context.Context, parent *zerolog.Logger) context.Context {
	subCtx := parent.WithContext(ctx)
	logger := log.Ctx(subCtx).With().Logger()
	return logger.WithContext(subCtx)
}

func StrToLogCtx(ctx context.Context, key string, value string) {
	l := zerolog.Ctx(ctx)
	l.UpdateContext(func(c zerolog.Context) zerolog.Context {
		return c.Str(key, value)
	})
}
