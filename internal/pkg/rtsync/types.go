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

package rtsync

import (
	"context"
	"time"
)

const (

	// used by one instance of ears to ask others to sync routing table

	EARS_REDIS_SYNC_CHANNEL = "ears_sync"

	// used by one instance of ears to tell all others that it just finished syncing its routing table

	EARS_REDIS_ACK_CHANNEL = "ears_ack"

	// used by all ears instances to do regular pings so we have awareness of how many instances are alive (could be done with redis api instead)

	EARS_REDIS_PING_CHANNEL = "ears_ping"

	// used by ears instance to tell other interested parties about updates (this can probably be removed)

	EARS_REDIS_MUTATION_CHANNEL = "ears_mutation"

	// operations

	EARS_REDIS_ADD_ROUTE_CMD = "add"

	EARS_REDIS_REMOVE_ROUTE_CMD = "rem"

	EARS_REDIS_RETRY_INTERVAL_SECONDS = 10 * time.Second

	EARS_DEFAULT_REDIS_ENDPOINT = "gears-redis-qa-001.6bteey.0001.usw2.cache.amazonaws.com:6379"
)

type RoutingTableSyncer interface {
	PublishSyncRequest(ctx context.Context, routeId string, add bool) error
	ListenForSyncRequests(ctx context.Context)
	PublishAckMessage(ctx context.Context) error
	PublishPings(ctx context.Context)
	ListenForPingMessages(ctx context.Context)
	GetInstanceCount(ctx context.Context) int
	PublishMutationMessage(ctx context.Context, routeId string, add bool) error
}
