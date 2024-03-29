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
	"github.com/xmidt-org/ears/internal/pkg/db"
	"github.com/xmidt-org/ears/internal/pkg/rtsemconv"
	"github.com/xmidt-org/ears/pkg/route"
	"github.com/xmidt-org/ears/pkg/tenant"
	"go.opentelemetry.io/otel/semconv/v1.4.0"
	"time"
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

//TODO: handle timestamps here

func NewRedisDbStorer(config Config, logger *zerolog.Logger) (*RedisDbStorer, error) {
	rdb := &RedisDbStorer{
		endpoint:  config.GetString("ears.storage.route.endpoint"),
		tableName: "routes",
		logger:    logger,
		config:    config,
	}
	rdb.client = redis.NewClient(&redis.Options{
		Addr:     rdb.endpoint,
		Password: "",
		DB:       0,
	})
	logger.Info().Msg("connected to redis storage layer at " + rdb.endpoint)
	return rdb, nil
}

func (d *RedisDbStorer) GetRoute(ctx context.Context, tid tenant.Id, id string) (route.Config, error) {
	_, span := db.CreateSpan(ctx, "getRoute", semconv.DBSystemRedis,
		semconv.DBConnectionStringKey.String(d.endpoint), rtsemconv.DBTable.String(d.tableName))
	defer span.End()

	r := route.Config{}
	result, err := d.client.HGet(d.tableName, tid.KeyWithRoute(id)).Result()
	if err != nil {
		if err.Error() == "redis: nil" {
			return r, &route.RouteNotFoundError{TenantId: tid, RouteId: id}
		}
		return r, fmt.Errorf("could not get route from redis: %v", err)
	}
	if result == "" {
		return r, &route.RouteNotFoundError{TenantId: tid, RouteId: id}
	}
	err = json.Unmarshal([]byte(result), &r)
	return r, err
}

func (d *RedisDbStorer) GetAllRoutes(ctx context.Context) ([]route.Config, error) {
	_, span := db.CreateSpan(ctx, "getRoutes", semconv.DBSystemRedis,
		semconv.DBConnectionStringKey.String(d.endpoint), rtsemconv.DBTable.String(d.tableName))
	defer span.End()
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

func (d *RedisDbStorer) GetAllTenantRoutes(ctx context.Context, tid tenant.Id) ([]route.Config, error) {
	_, span := db.CreateSpan(ctx, "getTenantRoutes", semconv.DBSystemRedis,
		semconv.DBConnectionStringKey.String(d.endpoint), rtsemconv.DBTable.String(d.tableName))
	defer span.End()
	routes, err := d.GetAllRoutes(ctx)
	if err != nil {
		return nil, err
	}
	filterRoutes := make([]route.Config, 0)
	for _, route := range routes {
		if route.TenantId.Equal(tid) {
			filterRoutes = append(filterRoutes, route)
		}
	}
	return filterRoutes, nil
}

func (d *RedisDbStorer) SetRoute(ctx context.Context, r route.Config) error {
	_, span := db.CreateSpan(ctx, "storeRoute", semconv.DBSystemRedis,
		semconv.DBConnectionStringKey.String(d.endpoint), rtsemconv.DBTable.String(d.tableName))
	defer span.End()
	if r.Id == "" {
		return fmt.Errorf("no route to store in redis")
	}
	r.Modified = time.Now().Unix()
	oldRoute, err := d.GetRoute(ctx, r.TenantId, r.Id)
	if err == nil {
		r.Created = oldRoute.Created
	} else {
		r.Created = r.Modified
	}
	val, err := json.Marshal(r)
	if err != nil {
		return err
	}
	_, err = d.client.HSet(d.tableName, r.TenantId.KeyWithRoute(r.Id), val).Result()
	if err != nil {
		return fmt.Errorf("could not insert route into redis: %v", err)
	}
	return nil
}

func (d *RedisDbStorer) SetRoutes(ctx context.Context, routes []route.Config) error {

	_, span := db.CreateSpan(ctx, "storeRoutes", semconv.DBSystemRedis,
		semconv.DBConnectionStringKey.String(d.endpoint), rtsemconv.DBTable.String(d.tableName))
	defer span.End()

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

func (d *RedisDbStorer) DeleteRoute(ctx context.Context, tid tenant.Id, id string) error {
	_, span := db.CreateSpan(ctx, "deleteRoute", semconv.DBSystemRedis,
		semconv.DBConnectionStringKey.String(d.endpoint), rtsemconv.DBTable.String(d.tableName))
	defer span.End()
	if id == "" {
		return fmt.Errorf("no route to delete in bolt")
	}
	_, err := d.client.HDel(d.tableName, tid.KeyWithRoute(id)).Result()
	if err != nil {
		return fmt.Errorf("could not delete route from redis: %v", err)
	}
	//if num != 1 {
	//return fmt.Errorf("could not delete route from redis")
	//}
	return nil
}

func (d *RedisDbStorer) DeleteRoutes(ctx context.Context, tid tenant.Id, ids []string) error {
	_, span := db.CreateSpan(ctx, "deleteRoutes", semconv.DBSystemRedis,
		semconv.DBConnectionStringKey.String(d.endpoint), rtsemconv.DBTable.String(d.tableName))
	defer span.End()
	for _, id := range ids {
		err := d.DeleteRoute(ctx, tid, id)
		if err != nil {
			return err
		}
	}
	return nil
}
