// Copyright 2020 Comcast Cable Communications Management, LLC
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

package app

import (
	"context"
	"errors"
	"fmt"
	"github.com/go-redis/redis"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"github.com/xmidt-org/ears/internal/pkg/logs"
	"os"
	"strings"
	"sync"
	"time"
)

type (
	RedisTableSyncer struct {
		sync.Mutex
		client          *redis.Client
		redisEndpoint   string
		routingTableMgr RoutingTableManager
		logger          *zerolog.Logger
		instanceId      string
	}
)

//DISCUSS: instance id
//DISCUSS: mocks
//DISCUSS: package structure

//TODO: add tests
//DONE: remove ping logic
//DONE: integrate with table manager
//TODO: load entire table on launch
//DONE: logging
//TODO: integrate with uberfx
//TODO: extract configs (redis endpoint etc.)
//TODO: start loading from dynamodb
//TODO: review notification payload (json)
//TODO: introduce global table hash
//TODO: consider app id and org id
//TODO: reload all when hash failure
//TODO: do not rely on hostname (alone)
//TODO: do we need to deal with multiple concurrent pub-ack handshakes?
//TODO: need local shared storage

func NewRedisTableSyncer(routingTableMgr RoutingTableManager, logger *zerolog.Logger) RoutingTableSyncer {
	endpoint := EARS_DEFAULT_REDIS_ENDPOINT
	s := new(RedisTableSyncer)
	s.logger = logger
	s.redisEndpoint = endpoint
	s.routingTableMgr = routingTableMgr
	s.client = redis.NewClient(&redis.Options{
		Addr:     endpoint,
		Password: "",
		DB:       0,
	})
	hostname, _ := os.Hostname()
	s.instanceId = hostname + "_" + uuid.New().String()
	s.ListenForSyncRequests()
	logger.Info().Msg("Launching Redis Syncer")
	return s
}

// PublishSyncRequest asks others to sync their routing tables

func (s *RedisTableSyncer) PublishSyncRequest(ctx context.Context, routeId string, add bool) error {
	// listen for ACKs first ...
	go func() {
		numSubscribers := s.GetInstanceCount(ctx)
		if numSubscribers <= 1 {
			s.logger.Info().Str("op", "PublishSyncRequest").Msg("no subscribers but me")
		} else {
			received := make(map[string]bool, 0)
			lrc := redis.NewClient(&redis.Options{
				Addr:     s.redisEndpoint,
				Password: "",
				DB:       0,
			})
			defer lrc.Close()
			pubsub := lrc.Subscribe(EARS_REDIS_ACK_CHANNEL)
			defer pubsub.Close()
			// 30 sec timeout on collecting acks
			done := make(chan bool, 1)
			go func() {
				for {
					msg, err := pubsub.ReceiveMessage()
					if err != nil {
						s.logger.Error().Str("op", "PublishSyncRequest").Msg(err.Error())
						break
					} else {
						s.logger.Info().Str("op", "PublishSyncRequest").Msg("receive ack on channel " + EARS_REDIS_ACK_CHANNEL)
					}
					received[msg.Payload] = true
					// wait until we received an ack from each subscriber (except the one originating the request)
					if len(received) >= numSubscribers-1 {
						break
					}
				}
				done <- true
			}()
			select {
			case <-done:
				s.logger.Info().Str("op", "PublishSyncRequest").Msg("done collecting acks")
			case <-time.After(30 * time.Second):
				s.logger.Info().Str("op", "PublishSyncRequest").Msg("timeout while collecting acks")
			}
		}
		s.PublishMutationMessage(ctx, routeId, add)
	}()
	// wait for listener to be ready
	time.Sleep(10 * time.Millisecond)
	// ... then request all flow apis to sync
	msg := ""
	if add {
		msg = EARS_REDIS_ADD_ROUTE_CMD
	} else {
		msg = EARS_REDIS_REMOVE_ROUTE_CMD
	}
	msg += "," + routeId + "," + s.instanceId
	err := s.client.Publish(EARS_REDIS_SYNC_CHANNEL, msg).Err()
	if err != nil {
		s.logger.Error().Str("op", "PublishSyncRequest").Msg(err.Error())
		return err
	} else {
		s.logger.Info().Str("op", "PublishSyncRequest").Msg("publish on channel " + EARS_REDIS_SYNC_CHANNEL)
	}
	return nil
}

// ListenForSyncRequests listens for sync request

func (s *RedisTableSyncer) ListenForSyncRequests() {
	go func() {
		ctx := context.Background()
		ctx = logs.SubLoggerCtx(ctx, s.logger)
		lrc := redis.NewClient(&redis.Options{
			Addr:     s.redisEndpoint,
			Password: "",
			DB:       0,
		})
		defer lrc.Close()
		pubsub := lrc.Subscribe(EARS_REDIS_SYNC_CHANNEL)
		defer pubsub.Close()
		for {
			msg, err := pubsub.ReceiveMessage()
			if err != nil {
				s.logger.Error().Str("op", "ListenForSyncRequests").Msg(err.Error())
				time.Sleep(EARS_REDIS_RETRY_INTERVAL_SECONDS)
			} else {
				s.logger.Info().Str("op", "ListenForSyncRequests").Msg("received message on channel " + EARS_REDIS_SYNC_CHANNEL)
				// parse message
				elems := strings.Split(msg.Payload, ",")
				if len(elems) != 3 {
					s.logger.Error().Str("op", "ListenForSyncRequests").Msg("bad message structure: " + msg.Payload)
				}
				// sync only whats needed
				if elems[2] != s.instanceId {
					if elems[0] == EARS_REDIS_ADD_ROUTE_CMD {
						s.logger.Info().Str("op", "ListenForSyncRequests").Str("instanceId", s.instanceId).Msg("received message to add route " + elems[1])
						err = s.routingTableMgr.SyncRouteAdded(ctx, elems[1])
					} else if elems[0] == EARS_REDIS_REMOVE_ROUTE_CMD {
						s.logger.Info().Str("op", "ListenForSyncRequests").Str("instanceId", s.instanceId).Msg("received message to remove route " + elems[1])
						err = s.routingTableMgr.SyncRouteRemoved(ctx, elems[1])
					} else {
						err = errors.New("bad command " + elems[0])
					}
					if err != nil {
						s.logger.Error().Str("op", "ListenForSyncRequests").Str("instanceId", s.instanceId).Msg(err.Error())
					}
					s.PublishAckMessage(ctx)
				} else {
					s.logger.Info().Str("op", "ListenForSyncRequests").Str("instanceId", s.instanceId).Msg("no need to sync myself")
				}
			}
		}
	}()
}

// PublishAckMessage confirm successful syncing of routing table

func (s *RedisTableSyncer) PublishAckMessage(ctx context.Context) error {
	err := s.client.Publish(EARS_REDIS_ACK_CHANNEL, s.instanceId).Err()
	if err != nil {
		s.logger.Error().Str("op", "PublishAckMessage").Msg(err.Error())
	} else {
		s.logger.Info().Str("op", "PublishAckMessage").Msg("published ack message")
	}
	return err
}

// GetSubscriberCount gets number of live ears instances

func (s *RedisTableSyncer) GetInstanceCount(ctx context.Context) int {
	channelMap, err := s.client.PubSubNumSub(EARS_REDIS_SYNC_CHANNEL).Result()
	if err != nil {
		s.logger.Error().Str("op", "GetInstanceCount").Msg(err.Error())
		return 0
	}
	numSubscribers, ok := channelMap[EARS_REDIS_SYNC_CHANNEL]
	if !ok {
		s.logger.Error().Str("op", "GetInstanceCount").Msg("could not get subscriber count for " + EARS_REDIS_SYNC_CHANNEL)
		return 0
	}
	s.logger.Debug().Str("op", "GetInstanceCount").Msg(fmt.Sprintf("num subscribers for channel %s is %d", EARS_REDIS_SYNC_CHANNEL, numSubscribers))
	return int(numSubscribers)
}

// PublishMutationMessage lets other interested parties know that updates are available

func (s *RedisTableSyncer) PublishMutationMessage(ctx context.Context, routeId string, add bool) error {
	//TODO: may not need this
	msg := ""
	if add {
		msg = EARS_REDIS_ADD_ROUTE_CMD
	} else {
		msg = EARS_REDIS_REMOVE_ROUTE_CMD
	}
	msg += "," + routeId + "," + s.instanceId
	s.logger.Info().Str("op", "ListenForPingMessages").Msg("published mutation message")
	return s.client.Publish(EARS_REDIS_MUTATION_CHANNEL, msg).Err()
}
