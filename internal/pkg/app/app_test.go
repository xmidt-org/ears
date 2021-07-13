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

package app_test

import (
	"errors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
	"github.com/xmidt-org/ears/internal/pkg/app"
	"github.com/xmidt-org/ears/internal/pkg/appsecret"
	"github.com/xmidt-org/ears/internal/pkg/config"
	"github.com/xmidt-org/ears/internal/pkg/fx/pluginmanagerfx"
	"github.com/xmidt-org/ears/internal/pkg/fx/quotamanagerfx"
	"github.com/xmidt-org/ears/internal/pkg/fx/routestorerfx"
	"github.com/xmidt-org/ears/internal/pkg/fx/syncerfx"
	"github.com/xmidt-org/ears/internal/pkg/fx/tenantstorerfx"
	"github.com/xmidt-org/ears/internal/pkg/tablemgr"
	testLog "github.com/xmidt-org/ears/test/log"
	"go.uber.org/fx"
	"os"
	"testing"
	"time"
)

func AppConfig() config.Config {
	v := viper.New()
	v.Set("ears.logLevel", "info")
	v.Set("ears.api.port", 8080)
	v.Set("ears.storage.route.type", "inmemory")
	v.Set("ears.storage.tenant.type", "inmemory")
	v.Set("ears.synchronization.type", "inmemory")
	v.Set("ears.ratelimiter.type", "inmemory")
	return v
}

func BadConfig() config.Config {
	v := viper.New()
	v.Set("ears.logLevel", "info")
	v.Set("ears.api.port", 0)
	v.Set("ears.storage.route.type", "inmemory")
	v.Set("ears.storage.tenant.type", "inmemory")
	v.Set("ears.synchronization.type", "inmemory")
	v.Set("ears.ratelimiter.type", "inmemory")
	return v
}

type NilLifeCycle struct {
}

func (lc *NilLifeCycle) Append(hook fx.Hook) {
	//do nothing
}

func TestSetupAPIServer(t *testing.T) {
	//Case 1: normal case
	err := app.SetupAPIServer(&NilLifeCycle{}, AppConfig(), &log.Logger, nil)
	if err != nil {
		t.Errorf("error setup API server %s", err.Error())
	}

	//Case 2: error case
	var invalidOptionErr *app.InvalidOptionError
	err = app.SetupAPIServer(&NilLifeCycle{}, BadConfig(), &log.Logger, nil)
	if err == nil || !errors.As(err, &invalidOptionErr) {
		t.Error("expect invalid optional error due to bad port")
	}
}

var applogger zerolog.Logger

func GetTestLogger() *zerolog.Logger {
	return &applogger
}

func TestAppRunSuccess(t *testing.T) {

	logListener := testLog.NewLogListener()
	applogger = zerolog.New(logListener)

	earsApp := fx.New(
		pluginmanagerfx.Module,
		routestorerfx.Module,
		syncerfx.Module,
		tenantstorerfx.Module,
		quotamanagerfx.Module,
		fx.Provide(
			AppConfig,
			appsecret.NewConfigVault,
			GetTestLogger,
			app.NewAPIManager,
			tablemgr.NewRoutingTableManager,
			app.NewMiddleware,
			app.NewMux,
		),
		fx.Invoke(quotamanagerfx.SetupQuotaManager),
		fx.Invoke(syncerfx.SetupDeltaSyncer),
		fx.Invoke(tablemgr.SetupRoutingManager),
		fx.Invoke(app.SetupAPIServer),
	)

	go func() {
		//cancel after we get the app startup success event
		//Server must be up in less than 10 seconds
		err := logListener.Listen("message", "API Server Started", 10*time.Second)
		if err != nil {
			t.Errorf("Fail to listen to log output, error=%s\n", err.Error())
		}
		//simulate interrupt signal
		pid := os.Getpid()
		p, _ := os.FindProcess(pid)
		p.Signal(os.Interrupt)
	}()

	earsApp.Run()
}
