// SPDX-FileCopyrightText: 2025 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

//go:build integration
// +build integration

package syncer_test

import (
	"testing"

	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
	"github.com/xmidt-org/ears/internal/pkg/config"
	"github.com/xmidt-org/ears/internal/pkg/syncer"
	"github.com/xmidt-org/ears/internal/pkg/syncer/redis"
)

func RedisConfig() config.Config {
	v := viper.New()
	v.Set("ears.synchronization.endpoint", "localhost:6379")
	v.Set("ears.synchronization.active", true)
	return v
}

func newRedisDeltaSyncer() syncer.DeltaSyncer {
	return redis.NewRedisDeltaSyncer(&log.Logger, RedisConfig())
}

func TestRedisDeltaSyncer(t *testing.T) {
	testSyncers(newRedisDeltaSyncer, t)
}
