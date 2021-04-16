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
	"github.com/xmidt-org/ears/internal/pkg/config"
	"github.com/xmidt-org/ears/internal/pkg/logs"
	"github.com/xmidt-org/ears/pkg/tenant"
	"strings"
	"sync"
)

var (
	lock                         = &sync.Mutex{}
	inmemoryDeltaSyncerSingleton *InmemoryDeltaSyncer
)

type (
	InmemoryDeltaSyncer struct {
		sync.Mutex
		notify            chan string
		active            bool
		instanceCnt       int
		localTableSyncers map[RoutingTableLocalSyncer]struct{}
		logger            *zerolog.Logger
		config            config.Config
	}
)

func NewInMemoryDeltaSyncer(logger *zerolog.Logger, config config.Config) RoutingTableDeltaSyncer {
	// This delta syncer is mainly for testing purposes. For it to work, multiple ears runtimes
	// should run within the same process and share the same instance of the in memory delta
	// syncer - we are forcing this here with a singleton
	lock.Lock()
	defer lock.Unlock()
	if inmemoryDeltaSyncerSingleton == nil {
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
		inmemoryDeltaSyncerSingleton = s
	}
	return inmemoryDeltaSyncerSingleton
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

func (s *InmemoryDeltaSyncer) PublishSyncRequest(ctx context.Context, tid tenant.Id, routeId string, instanceId string, add bool) {
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
			msg := cmd + "," + routeId + "," + instanceId + "," + sid + "," + tid.String()
			s.notify <- msg
		}()
	}
}

// StopListeningForSyncRequests stops listening for sync requests
func (s *InmemoryDeltaSyncer) StopListeningForSyncRequests(instanceId string) {
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
			if len(elems) != 5 {
				s.logger.Error().Str("op", "ListenForSyncRequests").Msg("bad message structure: " + msg)
			}
			// leave sync loop if asked
			if elems[0] == EARS_STOP_LISTENING_CMD {
				if elems[2] == instanceId || elems[2] == "" {
					s.logger.Info().Str("op", "ListenForSyncRequests").Str("instanceId", elems[2]).Msg("received stop listening message")
					return
				}
			}
			if elems[0] == EARS_ADD_ROUTE_CMD {
				s.logger.Info().Str("op", "ListenForSyncRequests").Str("instanceId", elems[2]).Str("routeId", elems[1]).Str("sid", elems[3]).Msg("received message to add route")

				tid, err := tenant.FromString(elems[4])
				if err != nil {
					s.logger.Error().Str("op", "ListenForSyncRequests").Str("instanceId", elems[2]).Str("routeId", elems[1]).Str("sid", elems[3]).Msg("failed to sync route: " + err.Error())
					continue
				}
				s.Lock()
				for localTableSyncer, _ := range s.localTableSyncers {
					if elems[2] != localTableSyncer.GetInstanceId() {
						err = localTableSyncer.SyncRoute(ctx, *tid, elems[1], true)
						if err != nil {
							s.logger.Error().Str("op", "ListenForSyncRequests").Str("instanceId", elems[2]).Str("routeId", elems[1]).Str("sid", elems[3]).Msg("failed to sync route: " + err.Error())
						}
					}
				}
				s.Unlock()
			} else if elems[0] == EARS_REMOVE_ROUTE_CMD {
				s.logger.Info().Str("op", "ListenForSyncRequests").Str("instanceId", elems[2]).Str("routeId", elems[1]).Str("sid", elems[3]).Msg("received message to remove route")

				tid, err := tenant.FromString(elems[4])
				if err != nil {
					s.logger.Error().Str("op", "ListenForSyncRequests").Str("instanceId", elems[2]).Str("routeId", elems[1]).Str("sid", elems[3]).Msg("failed to sync route: " + err.Error())
					continue
				}
				s.Lock()
				for localTableSyncer, _ := range s.localTableSyncers {
					if elems[2] != localTableSyncer.GetInstanceId() {
						err = localTableSyncer.SyncRoute(ctx, *tid, elems[1], false)
						if err != nil {
							s.logger.Error().Str("op", "ListenForSyncRequests").Str("instanceId", elems[2]).Str("routeId", elems[1]).Str("sid", elems[3]).Msg("failed to sync route: " + err.Error())
						}
					}
				}
				s.Unlock()
			} else if elems[0] == EARS_STOP_LISTENING_CMD {
				s.logger.Info().Str("op", "ListenForSyncRequests").Str("instanceId", elems[2]).Msg("stop message ignored")
				// already handled above
			} else {
				s.logger.Error().Str("op", "ListenForSyncRequests").Str("instanceId", elems[2]).Str("routeId", elems[1]).Str("sid", elems[3]).Msg("bad command " + elems[0])
			}
		}
	}()
}

// GetSubscriberCount gets number of live ears instances

func (s *InmemoryDeltaSyncer) GetInstanceCount(ctx context.Context) int {
	if !s.active {
		return 0
	}
	s.Lock()
	defer s.Unlock()
	return len(s.localTableSyncers)
}
