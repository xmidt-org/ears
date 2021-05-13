package syncer_test

import (
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
	"github.com/xmidt-org/ears/internal/pkg/config"
	"github.com/xmidt-org/ears/internal/pkg/syncer"
	"testing"
)

func InMemoryConfig() config.Config {
	v := viper.New()
	v.Set("ears.synchronization.active", true)
	return v
}

func newInMemoryDeltaSyncer() syncer.DeltaSyncer {
	return syncer.NewInMemoryDeltaSyncer(&log.Logger, InMemoryConfig())
}

func TestInMemoryDeltaSyncer(t *testing.T) {
	testSyncers(newInMemoryDeltaSyncer, t)
}
