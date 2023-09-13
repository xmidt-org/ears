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

package syncer

import (
	"context"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"github.com/xmidt-org/ears/internal/pkg/config"
	"github.com/xmidt-org/ears/pkg/tenant"
	"os"
	"sync"
)

var (
	syncerGroup = DeltaSyncerGroup{
		syncers: make(map[string]*InmemoryDeltaSyncer),
	}
	earsMetrics = make(map[string]*EarsMetric, 0)
)

type (
	DeltaSyncerGroup struct {
		sync.Mutex
		syncers map[string]*InmemoryDeltaSyncer
	}

	InmemoryDeltaSyncer struct {
		sync.Mutex
		active       bool
		instanceId   string
		localSyncers map[string][]LocalSyncer
		logger       *zerolog.Logger
	}
)

func NewInMemoryDeltaSyncer(logger *zerolog.Logger, config config.Config) DeltaSyncer {
	// This delta syncer is mainly for testing purposes. For it to work, groups of
	// in memory delta syncer will be able to see each other
	s := new(InmemoryDeltaSyncer)
	s.logger = logger
	s.localSyncers = make(map[string][]LocalSyncer)
	hostname, _ := os.Hostname()
	s.instanceId = hostname + "_" + uuid.New().String()
	s.active = config.GetBool("ears.synchronization.active")
	if !s.active {
		logger.Info().Msg("InMemory Delta Syncer Not Activated")
	}
	return s
}

func (s *InmemoryDeltaSyncer) DeleteMetrics(id string) {
	s.Lock()
	defer s.Unlock()
	delete(earsMetrics, id)
}

func (s *InmemoryDeltaSyncer) WriteMetrics(id string, metric *EarsMetric) {
	s.Lock()
	defer s.Unlock()
	earsMetrics[id] = metric
}

func (s *InmemoryDeltaSyncer) ReadMetrics(id string) *EarsMetric {
	s.Lock()
	defer s.Unlock()
	return earsMetrics[id]
}

func (s *InmemoryDeltaSyncer) RegisterLocalSyncer(itemType string, localSyncer LocalSyncer) {
	s.Lock()
	defer s.Unlock()
	s.localSyncers[itemType] = append(s.localSyncers[itemType], localSyncer)
}

func (s *InmemoryDeltaSyncer) UnregisterLocalSyncer(itemType string, localSyncer LocalSyncer) {
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

func (s *InmemoryDeltaSyncer) PublishSyncRequest(ctx context.Context, tid tenant.Id, itemType string, itemId string, add bool) {
	if !s.active {
		return
	}
	cmd := ""
	if add {
		cmd = EARS_ADD_ITEM_CMD
	} else {
		cmd = EARS_REMOVE_ITEM_CMD
	}
	sid := uuid.New().String() // session id
	numSubscribers := s.GetInstanceCount(ctx)
	if numSubscribers <= 1 {
		s.logger.Info().Str("op", "PublishSyncRequest").Msg("no subscribers but me - no need to publish sync")
	} else {
		syncerGroup.Lock()
		for id, syncer := range syncerGroup.syncers {
			if id == s.instanceId {
				continue
			}
			localSyncer := syncer
			go func() {
				msg := SyncCommand{
					cmd,
					itemType,
					itemId,
					s.instanceId,
					sid,
					tid,
				}

				localSyncer.notify(ctx, msg)
			}()
		}
		syncerGroup.Unlock()
	}
}

func (s *InmemoryDeltaSyncer) notify(ctx context.Context, msg SyncCommand) {
	if msg.Cmd == EARS_ADD_ITEM_CMD {
		s.logger.Info().Str("op", "ListenForSyncRequests").Str("instanceId", msg.InstanceId).Str("routeId", msg.ItemId).Str("sid", msg.Sid).Msg("received message to add route")

		s.Lock()
		syncers, ok := s.localSyncers[msg.ItemType]
		if ok {
			for _, localSyncer := range syncers {
				err := localSyncer.SyncItem(ctx, msg.Tenant, msg.ItemId, true)
				if err != nil {
					s.logger.Error().Str("op", "ListenForSyncRequests").Str("instanceId", msg.InstanceId).Str("routeId", msg.ItemId).Str("sid", msg.Sid).Msg("failed to sync route: " + err.Error())
				}
			}
		}
		s.Unlock()
	} else if msg.Cmd == EARS_REMOVE_ITEM_CMD {
		s.logger.Info().Str("op", "ListenForSyncRequests").Str("instanceId", msg.InstanceId).Str("routeId", msg.ItemId).Str("sid", msg.Sid).Msg("received message to remove route")

		s.Lock()
		syncers, ok := s.localSyncers[msg.ItemType]
		if ok {
			for _, localSyncer := range syncers {
				err := localSyncer.SyncItem(ctx, msg.Tenant, msg.ItemId, false)
				if err != nil {
					s.logger.Error().Str("op", "ListenForSyncRequests").Str("instanceId", msg.InstanceId).Str("routeId", msg.ItemId).Str("sid", msg.Sid).Msg("failed to sync route: " + err.Error())
				}
			}
		}
		s.Unlock()
	} else if msg.Cmd == EARS_STOP_LISTENING_CMD {
		s.logger.Info().Str("op", "ListenForSyncRequests").Str("instanceId", msg.InstanceId).Msg("stop message ignored")
		// already handled above
	} else {
		s.logger.Error().Str("op", "ListenForSyncRequests").Str("instanceId", msg.InstanceId).Str("routeId", msg.ItemId).Str("sid", msg.Sid).Msg("bad command " + msg.Cmd)
	}
}

// ListenForSyncRequests listens for sync request
func (s *InmemoryDeltaSyncer) StartListeningForSyncRequests() {
	syncerGroup.Lock()
	defer syncerGroup.Unlock()
	syncerGroup.syncers[s.instanceId] = s
}

// StopListeningForSyncRequests stops listening for sync requests
func (s *InmemoryDeltaSyncer) StopListeningForSyncRequests() {
	syncerGroup.Lock()
	defer syncerGroup.Unlock()
	delete(syncerGroup.syncers, s.instanceId)
}

// GetSubscriberCount gets number of live ears instances

func (s *InmemoryDeltaSyncer) GetInstanceCount(ctx context.Context) int {
	if !s.active {
		return 0
	}
	s.Lock()
	defer s.Unlock()
	return len(syncerGroup.syncers)
}
