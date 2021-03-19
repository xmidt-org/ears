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

package tablemgr

import (
	"context"
	"fmt"
	"github.com/go-redis/redis"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"github.com/xmidt-org/ears/internal/pkg/logs"
	"strings"
	"sync"
	"time"
)

const (

	// used by one instance of ears to ask others to sync routing table

	EARS_REDIS_SYNC_CHANNEL = "ears_sync"

	// used by one instance of ears to tell all others that it just finished syncing its routing table

	EARS_REDIS_ACK_CHANNEL = "ears_ack"

	EARS_REDIS_RETRY_INTERVAL_SECONDS = 10 * time.Second
)

type (
	RedisDeltaSyncer struct {
		sync.Mutex
		redisEndpoint     string
		active            bool
		client            *redis.Client
		localTableSyncers map[RoutingTableLocalSyncer]struct{}
		logger            *zerolog.Logger
		config            Config
	}
)

func NewRedisDeltaSyncer(logger *zerolog.Logger, config Config) RoutingTableDeltaSyncer {
	s := new(RedisDeltaSyncer)
	s.logger = logger
	s.config = config
	s.redisEndpoint = config.GetString("ears.synchronization.endpoint")
	s.localTableSyncers = make(map[RoutingTableLocalSyncer]struct{}, 0)
	s.active = config.GetBool("ears.synchronization.active")
	if !s.active {
		logger.Info().Msg("Redis Delta Syncer Not Activated")
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

func (s *RedisDeltaSyncer) RegisterLocalTableSyncer(localTableSyncer RoutingTableLocalSyncer) {
	// the redis implementation should only ever register itself
	s.Lock()
	defer s.Unlock()
	s.localTableSyncers[localTableSyncer] = struct{}{}
}

func (s *RedisDeltaSyncer) UnregisterLocalTableSyncer(localTableSyncer RoutingTableLocalSyncer) {
	s.Lock()
	defer s.Unlock()
	delete(s.localTableSyncers, localTableSyncer)
}

// PublishSyncRequest asks others to sync their routing tables

func (s *RedisDeltaSyncer) PublishSyncRequest(ctx context.Context, routeId string, instanceId string, add bool) {
	if !s.active {
		return
	}
	cmd := ""
	if add {
		cmd = EARS_ADD_ROUTE_CMD
	} else {
		cmd = EARS_REMOVE_ROUTE_CMD
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
							//s.logger.Info().Str("op", "PublishSyncRequest").Msg("receive ack on channel " + EARS_REDIS_ACK_CHANNEL)
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
							//s.logger.Info().Str("op", "PublishSyncRequest").Msg("ignoring unrelated ack: " + msg.Payload)
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
			msg := cmd + "," + routeId + "," + instanceId + "," + sid
			err := s.client.Publish(EARS_REDIS_SYNC_CHANNEL, msg).Err()
			if err != nil {
				s.logger.Error().Str("op", "PublishSyncRequest").Msg(err.Error())
			} else {
				//s.logger.Info().Str("op", "PublishSyncRequest").Msg("publish on channel " + EARS_REDIS_SYNC_CHANNEL)
			}
		}
	}()
}

// StopListeningForSyncRequests stops listening for sync requests
func (s *RedisDeltaSyncer) StopListeningForSyncRequests(instanceId string) {
	if !s.active {
		return
	}
	msg := EARS_STOP_LISTENING_CMD + ",," + instanceId + ","
	err := s.client.Publish(EARS_REDIS_SYNC_CHANNEL, msg).Err()
	if err != nil {
		s.logger.Error().Str("op", "StopListeningForSyncRequests").Msg(err.Error())
	}
}

// ListenForSyncRequests listens for sync request
func (s *RedisDeltaSyncer) StartListeningForSyncRequests(instanceId string) {
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
				continue
			} else {
				//s.logger.Info().Str("op", "ListenForSyncRequests").Msg("received message on channel " + EARS_REDIS_SYNC_CHANNEL)
				// parse message
				elems := strings.Split(msg.Payload, ",")
				if len(elems) != 4 {
					s.logger.Error().Str("op", "ListenForSyncRequests").Msg("bad message structure: " + msg.Payload)
					continue
				}
				// leave sync loop if asked
				if elems[0] == EARS_STOP_LISTENING_CMD {
					if elems[2] == instanceId || elems[2] == "" {
						s.logger.Info().Str("op", "ListenForSyncRequests").Str("instanceId", instanceId).Msg("received stop listening message")
						return
					}
				}
				// sync only whats needed
				if elems[2] != instanceId {
					s.GetInstanceCount(ctx) // just for logging
					if elems[0] == EARS_ADD_ROUTE_CMD {
						s.logger.Info().Str("op", "ListenForSyncRequests").Str("instanceId", instanceId).Str("routeId", elems[1]).Str("sid", elems[3]).Msg("received message to add route")
						for localTableSyncer, _ := range s.localTableSyncers {
							err = localTableSyncer.SyncRoute(ctx, elems[1], true)
							if err != nil {
								s.logger.Error().Str("op", "ListenForSyncRequests").Str("instanceId", instanceId).Str("routeId", elems[1]).Str("sid", elems[3]).Msg("failed to sync route: " + err.Error())
							}
						}
						s.publishAckMessage(ctx, elems[0], elems[1], instanceId, elems[3])
					} else if elems[0] == EARS_REMOVE_ROUTE_CMD {
						s.logger.Info().Str("op", "ListenForSyncRequests").Str("instanceId", instanceId).Str("routeId", elems[1]).Str("sid", elems[3]).Msg("received message to remove route")
						for localTableSyncer, _ := range s.localTableSyncers {
							err = localTableSyncer.SyncRoute(ctx, elems[1], false)
							if err != nil {
								s.logger.Error().Str("op", "ListenForSyncRequests").Str("instanceId", instanceId).Str("routeId", elems[1]).Str("sid", elems[3]).Msg("failed to sync route: " + err.Error())
							}
						}
						s.publishAckMessage(ctx, elems[0], elems[1], instanceId, elems[3])
					} else if elems[0] == EARS_STOP_LISTENING_CMD {
						s.logger.Info().Str("op", "ListenForSyncRequests").Str("instanceId", instanceId).Msg("stop message ignored")
						// already handled above
					} else {
						s.logger.Error().Str("op", "ListenForSyncRequests").Str("instanceId", instanceId).Str("routeId", elems[1]).Str("sid", elems[3]).Msg("bad command " + elems[0])
					}
				} else {
					s.logger.Info().Str("op", "ListenForSyncRequests").Str("instanceId", instanceId).Msg("no need to sync myself")
				}
			}
		}
	}()
}

// PublishAckMessage confirm successful syncing of routing table

func (s *RedisDeltaSyncer) publishAckMessage(ctx context.Context, cmd string, routeId string, instanceId string, sid string) error {
	if !s.active {
		return nil
	}
	msg := cmd + "," + routeId + "," + instanceId + "," + sid
	err := s.client.Publish(EARS_REDIS_ACK_CHANNEL, msg).Err()
	if err != nil {
		s.logger.Error().Str("op", "PublishAckMessage").Msg(err.Error())
	} else {
		//s.logger.Info().Str("op", "PublishAckMessage").Msg("published ack message")
	}
	return err
}

// GetSubscriberCount gets number of live ears instances

func (s *RedisDeltaSyncer) GetInstanceCount(ctx context.Context) int {
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
