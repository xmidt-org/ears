package app

import (
	"context"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
	"github.com/xmidt-org/ears/test/util"
	"testing"
	"time"
)

func TestAppRun(t *testing.T) {
	logListener := util.NewLogListener()
	log.Logger = zerolog.New(logListener)

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		//cancel after we get the app startup success event
		//Server must be up in less than 10 seconds
		err := logListener.Listen("message", "API Server started", 10*time.Second)
		if err != nil {
			t.Fatalf("Fail to listen to log output, error=%s\n", err.Error())
		}
		cancel()
	}()

	viper.Set("ears.logLevel", "info")
	viper.Set("ears.api.port", 8080)
	err := Run(ctx)
	if err != nil {
		t.Fatalf("Fail to run app, error=%s\n", err.Error())
	}
}
