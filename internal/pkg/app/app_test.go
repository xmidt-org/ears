package app

import (
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
	testLog "github.com/xmidt-org/ears/test/log"
	"go.uber.org/fx"
	"os"
	"testing"
	"time"
)

func AppTestConfig() Config {
	viper.Set("ears.logLevel", "info")
	viper.Set("ears.api.port", 8080)
	return viper.Sub("ears")
}

func LogListener() *zerolog.Logger {
	return &log.Logger
}

func TestAppRun(t *testing.T) {
	logListener := testLog.NewLogListener()
	log.Logger = zerolog.New(logListener)

	earsApp := fx.New(
		fx.Provide(
			AppTestConfig,
			LogListener,
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
			t.Fatalf("Fail to listen to log output, error=%s\n", err.Error())
		}
		//simulate interrupt signal
		pid := os.Getpid()
		p, _ := os.FindProcess(pid)
		p.Signal(os.Interrupt)
	}()

	earsApp.Run()
}
