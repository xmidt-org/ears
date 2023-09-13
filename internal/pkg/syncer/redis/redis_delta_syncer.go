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

package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-redis/redis"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"github.com/xmidt-org/ears/internal/pkg/config"
	"github.com/xmidt-org/ears/internal/pkg/syncer"
	"github.com/xmidt-org/ears/pkg/logs"
	"github.com/xmidt-org/ears/pkg/tenant"
	"os"
	"sync"
	"time"
)

const (
	// used by one instance of ears to ask others to sync routing table
	EARS_REDIS_SYNC_CHANNEL = "ears_sync"

	// used by one instance of ears to tell all others that it just finished syncing its routing table
	EARS_REDIS_ACK_CHANNEL = "ears_ack"

	EARS_METRICS = "ears_metrics"

	EARS_REDIS_RETRY_INTERVAL_SECONDS = 10 * time.Second
)

type (
	RedisDeltaSyncer struct {
		sync.Mutex
		redisEndpoint string
		active        bool
		client        *redis.Client
		localSyncers  map[string][]syncer.LocalSyncer
		logger        *zerolog.Logger
		config        config.Config
		instanceId    string
		env           string
	}
)

func NewRedisDeltaSyncer(logger *zerolog.Logger, config config.Config) syncer.DeltaSyncer {
	s := new(RedisDeltaSyncer)
	s.logger = logger
	s.config = config
	s.redisEndpoint = config.GetString("ears.synchronization.endpoint")
	s.env = config.GetString("ears.env")
	s.localSyncers = make(map[string][]syncer.LocalSyncer)
	hostname, _ := os.Hostname()
	s.instanceId = hostname + "_" + uuid.New().String()
	s.client = redis.NewClient(&redis.Options{
		Addr:     s.redisEndpoint,
		Password: "",
		DB:       0,
	})
	//logger.Info().Msg("Redis Syncer Started")
	return s
}

// id should be a unique plugin id
func (s *RedisDeltaSyncer) WriteMetrics(id string, metric *syncer.EarsMetric) {
	redisMapName := EARS_METRICS + "_" + s.env + "_" + id
	metric.Ts = time.Now().Unix()
	buf, err := json.Marshal(metric)
	if err != nil {
		s.logger.Error().Str("op", "WriteMetrics").Str("error", err.Error()).Msg("failed to write metrics")
		return
	}
	err = s.client.HSet(redisMapName, s.instanceId, string(buf)).Err()
	if err != nil {
		s.logger.Error().Str("op", "WriteMetrics").Str("mapKey", s.instanceId).Str("mapName", redisMapName).Str("error", err.Error()).Msg("failed to write metrics")
	}
}

// id should be a unique plugin id
func (s *RedisDeltaSyncer) ReadMetrics(id string) *syncer.EarsMetric {
	redisMapName := EARS_METRICS + "_" + s.env + "_" + id
	result := s.client.HGetAll(redisMapName)
	metricsMap, err := result.Result()
	if err != nil {
		s.logger.Error().Str("op", "ReadMetrics").Str("mapName", redisMapName).Str("error", err.Error()).Msg("failed to read metrics")
	}
	delKeys := make([]string, 0)
	var aggMetric syncer.EarsMetric
	aggMetric.Ts = time.Now().Unix()
	for id, v := range metricsMap {
		var metric syncer.EarsMetric
		err = json.Unmarshal([]byte(v), &metric)
		if err != nil {
			s.logger.Error().Str("op", "ReadMetrics").Str("mapName", redisMapName).Str("error", err.Error()).Msg("failed to parse metrics")
			continue
		}
		if aggMetric.Ts-metric.Ts > 5*60 {
			delKeys = append(delKeys, id)
		} else {
			aggMetric.SuccessCount += metric.SuccessCount
			aggMetric.ErrorCount += metric.ErrorCount
			aggMetric.FilterCount += metric.FilterCount
			aggMetric.SuccessVelocity += metric.SuccessVelocity
			aggMetric.ErrorVelocity += metric.ErrorVelocity
			aggMetric.FilterVelocity += metric.FilterVelocity
			if metric.LastEventTs > aggMetric.LastEventTs {
				aggMetric.LastEventTs = metric.LastEventTs
			}
		}
	}
	for _, id := range delKeys {
		err = s.client.HDel(redisMapName, id).Err()
		if err != nil {
			s.logger.Error().Str("op", "ReadMetrics").Str("mapKey", id).Str("mapName", redisMapName).Str("error", err.Error()).Msg("failed to delete old metric")
		}
	}
	return &aggMetric
}

func (s *RedisDeltaSyncer) RegisterLocalSyncer(itemType string, localSyncer syncer.LocalSyncer) {
	// the redis implementation should only ever register itself
	s.Lock()
	defer s.Unlock()
	s.localSyncers[itemType] = append(s.localSyncers[itemType], localSyncer)
}

func (s *RedisDeltaSyncer) UnregisterLocalSyncer(itemType string, localSyncer syncer.LocalSyncer) {
	s.Lock()
	defer s.Unlock()
	syncers, ok := s.localSyncers[itemType]
	if !ok {
		return
	}
	for i, syncer := range syncers {
		if syncer == localSyncer {
			//delete by copy the last element to the current pos and then
			//shorten the array by one
			syncers[i] = syncers[len(syncers)-1]
			syncers[len(syncers)-1] = nil
			syncers = syncers[:len(syncers)-1]
			s.localSyncers[itemType] = syncers
			return
		}
	}
}

// PublishSyncRequest asks others to sync their routing tables
func (s *RedisDeltaSyncer) PublishSyncRequest(ctx context.Context, tid tenant.Id, itemType string, itemId string, add bool) {
	if !s.active {
		return
	}
	cmd := ""
	if add {
		cmd = syncer.EARS_ADD_ITEM_CMD
	} else {
		cmd = syncer.EARS_REMOVE_ITEM_CMD
	}
	sid := uuid.New().String() // session id
	numSubscribers := s.GetInstanceCount(ctx)
	if numSubscribers <= 1 {
		s.logger.Info().Str("op", "PublishSyncRequest").Msg("no subscribers but me - no need to wait for ack")
		return
	}
	// outer go func is so that PublishSyncRequest returns immediately
	// this primarily cause issues when multi route unit tests share the same debug receiver
	// in practice this may not be an issue
	go func() {
		var wg sync.WaitGroup
		wg.Add(1)
		//listen for ACKs first ...
		go func() {
			received := make(map[string]bool)
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
						s.logger.Error().Str("op", "PublishSyncRequest").Err(err).Msg("Error collecting ack")
						break
					}
					var syncCmd syncer.SyncCommand
					err = json.Unmarshal([]byte(msg.Payload), &syncCmd)
					if err != nil {
						s.logger.Error().Str("op", "PublishSyncRequest").Str("error", err.Error()).Msg("bad ack message structure: " + msg.Payload)
						break
					}
					// only collect acks for this session
					if cmd == syncCmd.Cmd && itemType == syncCmd.ItemType && itemId == syncCmd.ItemId && syncCmd.Sid == sid && tid.Equal(syncCmd.Tenant) {
						received[syncCmd.InstanceId] = true
						// wait until we received an ack from each subscriber (except the one originating the request)
						if len(received) >= numSubscribers-1 {
							break
						}
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
			// at this point the delta has been fully synchronized - may want to publish something about that here
		}()
		if numSubscribers <= 1 {
			s.logger.Info().Str("op", "PublishSyncRequest").Msg("no subscribers but me - no need to publish sync")
		} else {
			// wait for listener to be ready
			wg.Wait()
			// ... then request all flow apis to sync
			syncCmd := &syncer.SyncCommand{
				Cmd:        cmd,
				ItemType:   itemType,
				ItemId:     itemId,
				InstanceId: s.instanceId,
				Sid:        sid,
				Tenant:     tid,
			}

			msg, _ := json.Marshal(syncCmd)
			err := s.client.Publish(EARS_REDIS_SYNC_CHANNEL, string(msg)).Err()
			if err != nil {
				s.logger.Error().Str("op", "PublishSyncRequest").Err(err).Msg("Fail to publish sync request")
			}
		}
	}()
}

// StopListeningForSyncRequests stops listening for sync requests
func (s *RedisDeltaSyncer) StopListeningForSyncRequests() {
	if !s.active {
		return
	}
	syncCmd := &syncer.SyncCommand{
		Cmd:        syncer.EARS_STOP_LISTENING_CMD,
		InstanceId: s.instanceId,
	}
	msg, _ := json.Marshal(syncCmd)
	err := s.client.Publish(EARS_REDIS_SYNC_CHANNEL, string(msg)).Err()
	if err != nil {
		s.logger.Error().Str("op", "StopListeningForSyncRequests").Msg(err.Error())
	}
}

// ListenForSyncRequests listens for sync request
func (s *RedisDeltaSyncer) StartListeningForSyncRequests() {
	if !s.active {
		return
	}
	subscribeComplete := sync.WaitGroup{}
	subscribeComplete.Add(1)
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

		subscribeComplete.Done()
		for {
			msg, err := pubsub.ReceiveMessage()
			if err != nil {
				s.logger.Error().Str("op", "ListenForSyncRequests").Msg(err.Error())
				time.Sleep(EARS_REDIS_RETRY_INTERVAL_SECONDS)
				continue
			} else {
				s.logger.Info().Str("op", "ListenForSyncRequests").Msg("received message on channel " + EARS_REDIS_SYNC_CHANNEL)
				// parse message
				var syncCmd syncer.SyncCommand
				err := json.Unmarshal([]byte(msg.Payload), &syncCmd)
				if err != nil {
					s.logger.Error().Str("op", "ListenForSyncRequests").Str("error", err.Error()).Msg("bad message structure: " + msg.Payload)
					continue
				}
				// leave sync loop if asked
				if syncCmd.Cmd == syncer.EARS_STOP_LISTENING_CMD {
					if syncCmd.InstanceId == s.instanceId || syncCmd.InstanceId == "" {
						s.logger.Info().Str("op", "ListenForSyncRequests").Str("instanceId", s.instanceId).Msg("received stop listening message")
						return
					}
				}
				// sync only whats needed
				if syncCmd.InstanceId != s.instanceId {
					s.GetInstanceCount(ctx) // just for logging
					if syncCmd.Cmd == syncer.EARS_ADD_ITEM_CMD {
						s.logger.Info().Str("op", "ListenForSyncRequests").Str("instanceId", s.instanceId).Str("routeId", syncCmd.ItemId).Str("sid", syncCmd.Sid).Msg("received message to add route")

						s.Lock()
						syncers, ok := s.localSyncers[syncCmd.ItemType]
						if ok {
							for _, localSyncer := range syncers {
								err = localSyncer.SyncItem(ctx, syncCmd.Tenant, syncCmd.ItemId, true)
								if err != nil {
									s.logger.Error().Str("op", "ListenForSyncRequests").Str("instanceId", s.instanceId).Str("routeId", syncCmd.ItemId).Str("sid", syncCmd.Sid).Msg("failed to sync route: " + err.Error())
								}
							}
						}
						s.Unlock()
						s.publishAckMessage(ctx, syncCmd.Cmd, syncCmd.ItemType, syncCmd.ItemId, syncCmd.Sid, syncCmd.Tenant)
					} else if syncCmd.Cmd == syncer.EARS_REMOVE_ITEM_CMD {
						s.logger.Info().Str("op", "ListenForSyncRequests").Str("instanceId", s.instanceId).Str("routeId", syncCmd.ItemId).Str("sid", syncCmd.Sid).Msg("received message to remove route")
						s.Lock()
						syncers, ok := s.localSyncers[syncCmd.ItemType]
						if ok {
							for _, localSyncer := range syncers {
								err = localSyncer.SyncItem(ctx, syncCmd.Tenant, syncCmd.ItemId, false)
								if err != nil {
									s.logger.Error().Str("op", "ListenForSyncRequests").Str("instanceId", s.instanceId).Str("routeId", syncCmd.ItemId).Str("sid", syncCmd.Sid).Msg("failed to sync route: " + err.Error())
								}
							}
						}
						s.Unlock()
						s.publishAckMessage(ctx, syncCmd.Cmd, syncCmd.ItemType, syncCmd.ItemId, syncCmd.Sid, syncCmd.Tenant)
					} else if syncCmd.Cmd == syncer.EARS_STOP_LISTENING_CMD {
						s.logger.Info().Str("op", "ListenForSyncRequests").Str("instanceId", s.instanceId).Msg("stop message ignored")
						// already handled above
					} else {
						s.logger.Error().Str("op", "ListenForSyncRequests").Str("instanceId", s.instanceId).Str("routeId", syncCmd.ItemId).Str("sid", syncCmd.Sid).Msg("bad command " + syncCmd.Cmd)
					}
				} else {
					s.logger.Info().Str("op", "ListenForSyncRequests").Str("instanceId", s.instanceId).Msg("no need to sync myself")
				}
			}
		}
	}()
	subscribeComplete.Wait()
}

// PublishAckMessage confirm successful syncing of routing table

func (s *RedisDeltaSyncer) publishAckMessage(ctx context.Context, cmd string, itemType string, itemId string, sid string, tid tenant.Id) error {
	if !s.active {
		return nil
	}
	syncCmd := &syncer.SyncCommand{
		Cmd:        cmd,
		ItemType:   itemType,
		ItemId:     itemId,
		InstanceId: s.instanceId,
		Sid:        sid,
		Tenant:     tid,
	}
	msg, _ := json.Marshal(syncCmd)
	err := s.client.Publish(EARS_REDIS_ACK_CHANNEL, string(msg)).Err()
	if err != nil {
		s.logger.Error().Str("op", "PublishAckMessage").Msg(err.Error())
	} else {
		s.logger.Info().Str("op", "PublishAckMessage").Msg("published ack message on " + EARS_REDIS_ACK_CHANNEL + ": " + string(msg))
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
