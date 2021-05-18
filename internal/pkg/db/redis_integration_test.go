// Copyright 2021 Comcast Cable Communications Management, LLC
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

// +build integration

package db_test

import (
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
	"github.com/xmidt-org/ears/internal/pkg/config"
	"github.com/xmidt-org/ears/internal/pkg/db/redis"
	"testing"
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
