package app

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
)

func AppConfig() Config {
	v := viper.New()
	v.Set("ears.logLevel", "info")
	v.Set("ears.api.port", 8080)
	v.Set("ears.env", "test")
	return v.Sub("ears")
}

func BadConfig() Config {
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
	err := SetupAPIServer(&NilLifeCycle{}, AppConfig(), NewLogger(), nil)
	if err != nil {
		t.Errorf("error setup API server %s", err.Error())
	}

	//Case 2: error case
	var invalidOptionErr *InvalidOptionError
	err = SetupAPIServer(&NilLifeCycle{}, BadConfig(), NewLogger(), nil)
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
			NewLogger,
			NewRouter,
			NewMiddlewares,
			NewMux,
		),
		fx.Invoke(SetupAPIServer),
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
