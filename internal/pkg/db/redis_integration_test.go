// SPDX-FileCopyrightText: 2025 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

//go:build integration
// +build integration

package db_test

import (
	"testing"

	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
	"github.com/xmidt-org/ears/internal/pkg/config"
	"github.com/xmidt-org/ears/internal/pkg/db/redis"
)

func redisConfig() config.Config {
	v := viper.New()
	v.Set("ears.storage.route.endpoint", "127.0.0.1:6379")
	return v
}

func TestRedisRouteStorer(t *testing.T) {
	s, err := redis.NewRedisDbStorer(redisConfig(), &log.Logger)
	if err != nil {
		t.Fatalf("Error instantiate redisdb %s\n", err.Error())
	}
	testRouteStorer(s, t)
}
