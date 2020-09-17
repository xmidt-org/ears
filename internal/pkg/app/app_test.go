package app_test

import (
	"errors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
	testLog "github.com/xmidt-org/ears/test/log"
	"go.uber.org/fx"
	"os"
	"testing"
	"time"
	"github.com/xmidt-org/ears/internal/pkg/app"
)

func AppConfig() app.Config {
	v := viper.New()
	v.Set("ears.logLevel", "info")
	v.Set("ears.api.port", 8080)
	v.Set("ears.env", "test")
	return v.Sub("ears")
}

func BadConfig() app.Config {
	v := viper.New()
	v.Set("ears.logLevel", "info")
	v.Set("ears.api.port", 0)
	return v.Sub("ears")
}

type NilLifeCycle struct {
}

func (lc *NilLifeCycle) Append(hook fx.Hook) {
	//do nothing
}

func TestSetupAPIServer(t *testing.T) {
	//Case 1: normal case
	err := app.SetupAPIServer(&NilLifeCycle{}, AppConfig(), app.NewLogger(), nil)
	if err != nil {
		t.Errorf("error setup API server %s", err.Error())
	}

	//Case 2: error case
	var invalidOptionErr *app.InvalidOptionError
	err = app.SetupAPIServer(&NilLifeCycle{}, BadConfig(), app.NewLogger(), nil)
	if err == nil || !errors.As(err, &invalidOptionErr) {
		t.Error("expect invalid optional error due to bad config")
	}
}

func TestAppRunSuccess(t *testing.T) {

	logListener := testLog.NewLogListener()
	log.Logger = zerolog.New(logListener)

	earsApp := fx.New(
		fx.Provide(
			AppConfig,
			app.NewLogger,
			app.NewRouter,
			app.NewMiddlewares,
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
