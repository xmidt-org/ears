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
	testLog "github.com/xmidt-org/ears/test/log"
	"go.uber.org/fx"
	"os"
	"testing"
	"time"
)

func AppConfig() app.Config {
	v := viper.New()
	v.Set("ears.logLevel", "info")
	v.Set("ears.api.port", 8080)
	return v
}

func BadConfig() app.Config {
	v := viper.New()
	v.Set("ears.logLevel", "info")
	v.Set("ears.api.port", 0)
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
		t.Error("expect invalid optional error due to bad config")
	}
}

var applogger zerolog.Logger

func GetTestLogger() *zerolog.Logger {
	return &applogger
}

func TestAppRunSuccess(t *testing.T) {

	logListener := testLog.NewLogListener()
	applogger = zerolog.New(logListener)

	err := app.InitLogger(AppConfig())
	if err != nil {
		t.Errorf("Fail to initLogger: %s\n", err.Error())
	}
	earsApp := fx.New(
		fx.Provide(
			AppConfig,
			GetTestLogger,
			app.NewAPIManager,
			app.NewRoutingTableManager,
			app.NewMiddleware,
			app.NewMux,
		),
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
