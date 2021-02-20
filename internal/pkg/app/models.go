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
	"github.com/xmidt-org/ears/pkg/filter"
	"github.com/xmidt-org/ears/pkg/receiver"
	"github.com/xmidt-org/ears/pkg/sender"
	"time"

	"github.com/xmidt-org/ears/pkg/route"
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

type (

	// A RoutingTableManager supports modifying and querying an EARS routing table
	RoutingTableManager interface {
		// AddRoute adds a route to live routing table and runs it and also stores the route in the persistence layer
		AddRoute(ctx context.Context, route *route.Config) error
		// RemoveRoute removes a route from a live routing table and stops it and also removes the route from the persistence layer
		RemoveRoute(ctx context.Context, routeId string) error
		// GetRoute gets a single route by its ID
		GetRoute(ctx context.Context, routeId string) (*route.Config, error)
		// GetAllRoutes gets all live routes from persistence layer
		GetAllRoutes(ctx context.Context) ([]route.Config, error)
		// GetAllSenders gets all senders currently present in the system
		GetAllSenders(ctx context.Context) (map[string]sender.Sender, error)
		// GetAllReceivers gets all receivers currently present in the system
		GetAllReceivers(ctx context.Context) (map[string]receiver.Receiver, error)
		// GetAllFilters gets all filters currently present in the system
		GetAllFilters(ctx context.Context) (map[string]filter.Filterer, error)
		// SyncRouteAdded
		SyncRouteAdded(ctx context.Context, routeId string) error
		// SyncRouteRemoved
		SyncRouteRemoved(ctx context.Context, routeId string) error
		// SyncAllRoutes
		SyncAllRoutes(ctx context.Context) error
	}

	RoutingTableSyncer interface {
		PublishSyncRequest(ctx context.Context, routeId string, add bool) error
		ListenForSyncRequests(ctx context.Context)
		PublishAckMessage(ctx context.Context) error
		PublishPings(ctx context.Context)
		ListenForPingMessages(ctx context.Context)
		GetInstanceCount(ctx context.Context) int
		PublishMutationMessage(ctx context.Context, routeId string, add bool) error
	}
)
