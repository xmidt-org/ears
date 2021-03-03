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
	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"github.com/xmidt-org/ears/internal/pkg/logs"
	"strings"
	"sync"
)

type (
	InmemoryDeltaSyncer struct {
		sync.Mutex
		notify            chan string
		active            bool
		instanceCnt       int
		localTableSyncers map[RoutingTableLocalSyncer]struct{}
		logger            *zerolog.Logger
		config            Config
	}
)

func NewInMemoryDeltaSyncer(logger *zerolog.Logger, config Config) RoutingTableDeltaSyncer {
	// This delta syncer is mainly for testing purposes. For it to work, multiple ears runtimes
	// should run within the same process and share the same instance of the in memory delta
	// syncer.
	s := new(InmemoryDeltaSyncer)
	s.logger = logger
	s.config = config
	s.localTableSyncers = make(map[RoutingTableLocalSyncer]struct{}, 0)
	s.instanceCnt = 0
	s.active = config.GetBool("ears.synchronization.active")
	if !s.active {
		logger.Info().Msg("InMemory Delta Syncer Not Activated")
	} else {
		s.notify = make(chan string, 0)
	}
	return s
}

func (s *InmemoryDeltaSyncer) RegisterLocalTableSyncer(localTableSyncer RoutingTableLocalSyncer) {
	s.Lock()
	defer s.Unlock()
	s.localTableSyncers[localTableSyncer] = struct{}{}
}

func (s *InmemoryDeltaSyncer) UnregisterLocalTableSyncer(localTableSyncer RoutingTableLocalSyncer) {
	s.Lock()
	defer s.Unlock()
	delete(s.localTableSyncers, localTableSyncer)
}

// PublishSyncRequest asks others to sync their routing tables

func (s *InmemoryDeltaSyncer) PublishSyncRequest(ctx context.Context, routeId string, instanceId string, add bool) {
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
	if numSubscribers <= 1 {
		s.logger.Info().Str("op", "PublishSyncRequest").Msg("no subscribers but me - no need to publish sync")
	} else {
		go func() {
			msg := cmd + "," + routeId + "," + instanceId + "," + sid
			s.notify <- msg
		}()
	}
}

// StopListeningForSyncRequests stops listening for sync requests
func (s *InmemoryDeltaSyncer) StopListeningForSyncRequests(instanceId string) {
	if !s.active {
		return
	}
}

// ListenForSyncRequests listens for sync request
func (s *InmemoryDeltaSyncer) StartListeningForSyncRequests(instanceId string) {
	if !s.active {
		return
	}
	go func() {
		ctx := context.Background()
		ctx = logs.SubLoggerCtx(ctx, s.logger)
		for msg := range s.notify {
			elems := strings.Split(msg, ",")
			if len(elems) != 4 {
				s.logger.Error().Str("op", "ListenForSyncRequests").Msg("bad message structure: " + msg)
			}
			// leave sync loop if asked
			if elems[0] == EARS_STOP_LISTENING_CMD {
				if elems[2] == instanceId || elems[2] == "" {
					s.logger.Info().Str("op", "ListenForSyncRequests").Str("instanceId", instanceId).Msg("received stop listening message")
					return
				}
			}
			// sync only whats needed
			if elems[2] != instanceId { // need to manage instance IDs better
				s.GetInstanceCount(ctx) // just for logging
				var err error
				if elems[0] == EARS_ADD_ROUTE_CMD {
					s.logger.Info().Str("op", "ListenForSyncRequests").Str("instanceId", instanceId).Str("routeId", elems[1]).Str("sid", elems[3]).Msg("received message to add route")
					for localTableSyncer, _ := range s.localTableSyncers {
						err = localTableSyncer.SyncRoute(ctx, elems[1], true)
						if err != nil {
							s.logger.Error().Str("op", "ListenForSyncRequests").Str("instanceId", instanceId).Str("routeId", elems[1]).Str("sid", elems[3]).Msg("failed to sync route: " + err.Error())
						}
					}
				} else if elems[0] == EARS_REMOVE_ROUTE_CMD {
					s.logger.Info().Str("op", "ListenForSyncRequests").Str("instanceId", instanceId).Str("routeId", elems[1]).Str("sid", elems[3]).Msg("received message to remove route")
					for localTableSyncer, _ := range s.localTableSyncers {
						err = localTableSyncer.SyncRoute(ctx, elems[1], false)
						if err != nil {
							s.logger.Error().Str("op", "ListenForSyncRequests").Str("instanceId", instanceId).Str("routeId", elems[1]).Str("sid", elems[3]).Msg("failed to sync route: " + err.Error())
						}
					}
				} else if elems[0] == EARS_STOP_LISTENING_CMD {
					s.logger.Info().Str("op", "ListenForSyncRequests").Str("instanceId", instanceId).Msg("stop message ignored")
					// already handled above
				} else {
					s.logger.Error().Str("op", "ListenForSyncRequests").Str("instanceId", instanceId).Str("routeId", elems[1]).Str("sid", elems[3]).Msg("bad command " + elems[0])
				}
			} else {
				s.logger.Info().Str("op", "ListenForSyncRequests").Str("instanceId", instanceId).Msg("no need to sync just myself")
			}
		}
	}()
}

// GetSubscriberCount gets number of live ears instances

func (s *InmemoryDeltaSyncer) GetInstanceCount(ctx context.Context) int {
	if !s.active {
		return 0
	}
	return len(s.localTableSyncers)
}
