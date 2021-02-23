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
		redisEndpoint   string
		active          bool
		instanceId      string
		client          *redis.Client
		routingTableMgr RoutingTableManager
		logger          *zerolog.Logger
		config          Config
	}
)

//DISCUSS: instance id
//DISCUSS: mocks
//DISCUSS: package structure

//TODO: ensure all unit tests pass
//TODO: integrate with uberfx
//TODO: do we need to deal with multiple concurrent pub-ack handshakes?
//TODO: unsubscribe from channels on shutdown
//TODO: load entire table on launch
//TODO: shutdown on catastrophic failure
//
//TODO: introduce global table hash
//TODO: reload all when hash failure
//
//TODO: add tests
//TODO: review notification payload (json)
//TODO: consider app id and org id
//
//TODO: architecture.md
//
//DONE: remove ping logic
//DONE: integrate with table manager
//DONE: logging
//DONE: extract configs (redis endpoint etc.)
//DONE: do not rely on hostname (alone)
//DONE: need local shared storage (boltdb, redis)
//DONE: ability to turn syncing on and off

func NewRedisTableSyncer(routingTableMgr RoutingTableManager, logger *zerolog.Logger, config Config) RoutingTableSyncer {
	s := new(RedisTableSyncer)
	s.logger = logger
	s.config = config
	s.redisEndpoint = config.GetString("ears.synchronization.endpoint")
	s.routingTableMgr = routingTableMgr
	hostname, _ := os.Hostname()
	s.instanceId = hostname + "_" + uuid.New().String()
	s.active = config.GetBool("ears.synchronization.active")
	if !s.active {
		logger.Info().Msg("Redis Syncer Not Activated")
	} else {
		s.client = redis.NewClient(&redis.Options{
			Addr:     s.redisEndpoint,
			Password: "",
			DB:       0,
		})
		s.ListenForSyncRequests()
		logger.Info().Msg("Redis Syncer Started")
	}
	return s
}

// PublishSyncRequest asks others to sync their routing tables

func (s *RedisTableSyncer) PublishSyncRequest(ctx context.Context, routeId string, add bool) error {
	if !s.active {
		return nil
	}
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
	if !s.active {
		return
	}
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
	if !s.active {
		return nil
	}
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
	if !s.active {
		return 0
	}
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
	if !s.active {
		return nil
	}
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
