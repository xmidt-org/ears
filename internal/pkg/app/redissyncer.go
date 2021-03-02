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
//
//TODO: add tests
//TODO: consider app id and org id
//TODO: architecture.md
//
//DONE: reload all when hash failure or incremental repair
//DONE: consider introduce global table hash (rejetced)
//DONE: load entire table on launch, unload on stop
//DONE: deal with multiple concurrent pub-ack handshakes
//DONE: don't do acks (rejected)
//DONE: contain route config in sync message
//DONE: consider using one channel per transaction (rejected)
//DONE: ensure all unit tests pass (issue of concurrent execution remains for shared data stores)
//
//DONE: review notification payload (json)
//
//DONE: unsubscribe from channels on shutdown
//DONE: integrate with uberfx
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
		//s.ListenForSyncRequests()
		//logger.Info().Msg("Redis Syncer Started")
	}
	return s
}

// PublishSyncRequest asks others to sync their routing tables

func (s *RedisTableSyncer) PublishSyncRequest(ctx context.Context, routeId string, add bool) {
	if !s.active {
		return
	}
	cmd := ""
	if add {
		cmd = EARS_REDIS_ADD_ROUTE_CMD
	} else {
		cmd = EARS_REDIS_REMOVE_ROUTE_CMD
	}
	sid := uuid.New().String() // session id
	numSubscribers := s.GetInstanceCount(ctx)
	// outer go func is so that PublishSyncRequest returns immediately despite the 10 ms wait below
	// this primarily cause issues when multi route unit tests share the same debug receiver
	// in practice this may not be an issue
	go func() {
		var wg sync.WaitGroup
		wg.Add(1)
		// listen for ACKs first ...
		go func() {
			if numSubscribers <= 1 {
				s.logger.Info().Str("op", "PublishSyncRequest").Msg("no subscribers but me - no need to wait for ack")
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
				wg.Done()
				go func() {
					for {
						msg, err := pubsub.ReceiveMessage()
						if err != nil {
							s.logger.Error().Str("op", "PublishSyncRequest").Msg(err.Error())
							break
						} else {
							s.logger.Info().Str("op", "PublishSyncRequest").Msg("receive ack on channel " + EARS_REDIS_ACK_CHANNEL)
						}
						elems := strings.Split(msg.Payload, ",")
						if len(elems) != 4 {
							s.logger.Error().Str("op", "PublishSyncRequest").Msg("bad ack message structure: " + msg.Payload)
							break
						}
						// only collect acks for this session
						if cmd == elems[0] && routeId == elems[1] && elems[3] == sid {
							received[elems[2]] = true
							// wait until we received an ack from each subscriber (except the one originating the request)
							if len(received) >= numSubscribers-1 {
								break
							}
						} else {
							s.logger.Info().Str("op", "PublishSyncRequest").Msg("ignoring unrelated ack: " + msg.Payload)
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
			// at this point the delta has been fully synchronized - may want to publish something about that here
		}()
		if numSubscribers <= 1 {
			s.logger.Info().Str("op", "PublishSyncRequest").Msg("no subscribers but me - no need to publish sync")
		} else {
			// wait for listener to be ready
			//time.Sleep(10 * time.Millisecond)
			wg.Wait()
			// ... then request all flow apis to sync
			msg := cmd + "," + routeId + "," + s.instanceId + "," + sid
			err := s.client.Publish(EARS_REDIS_SYNC_CHANNEL, msg).Err()
			if err != nil {
				s.logger.Error().Str("op", "PublishSyncRequest").Msg(err.Error())
			} else {
				s.logger.Info().Str("op", "PublishSyncRequest").Msg("publish on channel " + EARS_REDIS_SYNC_CHANNEL)
			}
		}
	}()
}

// StopListeningForSyncRequests stops listening for sync requests
func (s *RedisTableSyncer) StopListeningForSyncRequests() {
	if !s.active {
		return
	}
	msg := EARS_REDIS_STOP_LISTENING_CMD + ",," + s.instanceId + ","
	err := s.client.Publish(EARS_REDIS_SYNC_CHANNEL, msg).Err()
	if err != nil {
		s.logger.Error().Str("op", "StopListeningForSyncRequests").Msg(err.Error())
	}
}

// ListenForSyncRequests listens for sync request
func (s *RedisTableSyncer) StartListeningForSyncRequests() {
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
				if len(elems) != 4 {
					s.logger.Error().Str("op", "ListenForSyncRequests").Msg("bad message structure: " + msg.Payload)
				}
				// leave sync loop if asked
				if elems[0] == EARS_REDIS_STOP_LISTENING_CMD {
					if elems[2] == s.instanceId || elems[2] == "" {
						s.logger.Info().Str("op", "ListenForSyncRequests").Str("instanceId", s.instanceId).Msg("received stop listening message")
						return
					}
				}
				// sync only whats needed
				if elems[2] != s.instanceId {
					if elems[0] == EARS_REDIS_ADD_ROUTE_CMD {
						s.logger.Info().Str("op", "ListenForSyncRequests").Str("instanceId", s.instanceId).Str("routeId", elems[1]).Str("sid", elems[3]).Msg("received message to add route")
						err = s.routingTableMgr.SyncRouteAdded(ctx, elems[1])
						s.PublishAckMessage(ctx, elems[0], elems[1], elems[2], elems[3])
					} else if elems[0] == EARS_REDIS_REMOVE_ROUTE_CMD {
						s.logger.Info().Str("op", "ListenForSyncRequests").Str("instanceId", s.instanceId).Str("routeId", elems[1]).Str("sid", elems[3]).Msg("received message to remove route")
						err = s.routingTableMgr.SyncRouteRemoved(ctx, elems[1])
						s.PublishAckMessage(ctx, elems[0], elems[1], elems[2], elems[3])
					} else if elems[0] == EARS_REDIS_STOP_LISTENING_CMD {
						s.logger.Info().Str("op", "ListenForSyncRequests").Str("instanceId", s.instanceId).Msg("stop message ignored")
						// already handled above
					} else {
						err = errors.New("bad command " + elems[0])
					}
					if err != nil {
						s.logger.Error().Str("op", "ListenForSyncRequests").Str("instanceId", s.instanceId).Msg(err.Error())
					}
				} else {
					s.logger.Info().Str("op", "ListenForSyncRequests").Str("instanceId", s.instanceId).Msg("no need to sync myself")
				}
			}
		}
	}()
}

// PublishAckMessage confirm successful syncing of routing table

func (s *RedisTableSyncer) PublishAckMessage(ctx context.Context, cmd string, routeId string, instanceId string, sid string) error {
	if !s.active {
		return nil
	}
	msg := cmd + "," + routeId + "," + instanceId + "," + sid
	err := s.client.Publish(EARS_REDIS_ACK_CHANNEL, msg).Err()
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
