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

package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-redis/redis"
	"github.com/rs/zerolog"
	"github.com/xmidt-org/ears/pkg/route"
)

type RedisDbStorer struct {
	client    *redis.Client
	endpoint  string
	tableName string
	logger    *zerolog.Logger
	config    Config
}

type Config interface {
	GetString(key string) string
	GetInt(key string) int
	GetBool(key string) bool
}

func NewRedisDbStorer(config Config, logger *zerolog.Logger) (*RedisDbStorer, error) {
	rdb := &RedisDbStorer{
		endpoint:  config.GetString("ears.storage.endpoint"),
		tableName: "routes",
		logger:    logger,
		config:    config,
	}
	rdb.client = redis.NewClient(&redis.Options{
		Addr:     rdb.endpoint,
		Password: "",
		DB:       0,
	})
	logger.Info().Msg("connected to redis at " + rdb.endpoint)
	return rdb, nil
}

func (d *RedisDbStorer) GetRoute(ctx context.Context, id string) (route.Config, error) {
	route := route.Config{}
	result, err := d.client.HGet(d.tableName, id).Result()
	if err != nil {
		return route, fmt.Errorf("could not get route from redis: %v", err)
	}
	if result == "" {
		return route, fmt.Errorf("could not get route from redis")
	}
	err = json.Unmarshal([]byte(result), &route)
	return route, err
}

func (d *RedisDbStorer) GetAllRoutes(ctx context.Context) ([]route.Config, error) {
	routes := make([]route.Config, 0)
	results, err := d.client.HGetAll(d.tableName).Result()
	if err != nil {
		return routes, fmt.Errorf("could not get routes from redis: %v", err)
	}
	for _, v := range results {
		var route route.Config
		err = json.Unmarshal([]byte(v), &route)
		if err != nil {
			return routes, err
		}
		routes = append(routes, route)
	}
	return routes, nil
}

func (d *RedisDbStorer) SetRoute(ctx context.Context, r route.Config) error {
	if r.Id == "" {
		return fmt.Errorf("no route to store in redis")
	}
	val, err := json.Marshal(r)
	if err != nil {
		return err
	}
	success, err := d.client.HSet(d.tableName, r.Id, val).Result()
	if err != nil {
		return fmt.Errorf("could not insert route into redis: %v", err)
	}
	if !success {
		return fmt.Errorf("could not insert route into redis")
	}
	return nil
}

func (d *RedisDbStorer) SetRoutes(ctx context.Context, routes []route.Config) error {
	if routes == nil {
		return fmt.Errorf("no routes to store in bolt")
	}
	for _, r := range routes {
		err := d.SetRoute(ctx, r)
		if err != nil {
			return err
		}
	}
	return nil
}

func (d *RedisDbStorer) DeleteRoute(ctx context.Context, id string) error {
	if id == "" {
		return fmt.Errorf("no route to delete in bolt")
	}

	num, err := d.client.HDel(d.tableName, id).Result()
	if err != nil {
		return fmt.Errorf("could not delete route from redis: %v", err)
	}
	if num != 1 {
		return fmt.Errorf("could not delete route from redis")
	}
	return nil
}

func (d *RedisDbStorer) DeleteRoutes(ctx context.Context, ids []string) error {
	for _, id := range ids {
		err := d.DeleteRoute(ctx, id)
		if err != nil {
			return err
		}
	}
	return nil
}
