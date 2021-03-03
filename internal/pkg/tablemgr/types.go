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
	"github.com/xmidt-org/ears/pkg/filter"
	"github.com/xmidt-org/ears/pkg/receiver"
	"github.com/xmidt-org/ears/pkg/route"
	"github.com/xmidt-org/ears/pkg/sender"
)

const (
	EARS_ADD_ROUTE_CMD = "add"

	EARS_REMOVE_ROUTE_CMD = "rem"

	EARS_STOP_LISTENING_CMD = "stop"
)

type Config interface {
	GetString(key string) string
	GetInt(key string) int
	GetBool(key string) bool
}

type (

	// A RoutingTableManager supports modifying and querying an EARS routing table
	RoutingTableManager interface {
		RoutingTableDeltaSyncer  // routing table manager delegates to routing table delta syncer for fanning out changes
		RoutingTableGlobalSyncer // routing table manager delegates to routing table global syncer for startup and tear down
		RoutingTableLocalSyncer  // to sync routing table upon receipt of an update notification for a single route
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
		// GetInstanceId
		GetInstanceId() string
	}

	RoutingTableGlobalSyncer interface {
		// StartGlobalSyncChecker
		StartGlobalSyncChecker()
		// RegisterAllRoutes
		RegisterAllRoutes() error
		// UnregisterAllRoutes
		UnregisterAllRoutes() error
		// SynchronizeAllRoutes
		SynchronizeAllRoutes() (int, error)
		// IsSynchronized
		IsSynchronized() (bool, error)
	}

	RoutingTableLocalSyncer interface {
		// SyncRoute
		SyncRoute(ctx context.Context, routeId string, add bool) error
		// GetInstanceId
		GetInstanceId() string
	}

	RoutingTableDeltaSyncer interface {
		// StartListeningForSyncRequests
		StartListeningForSyncRequests(instanceId string) // do we need this still?
		// StopListeningForSyncRequests
		StopListeningForSyncRequests(instanceId string) // do we need this still?
		// RegisterLocalTableSyncer
		RegisterLocalTableSyncer(localTableSyncer RoutingTableLocalSyncer)
		// UnregisterLocalTableSyncer
		UnregisterLocalTableSyncer(localTableSyncer RoutingTableLocalSyncer)
		// PublishSyncRequest
		PublishSyncRequest(ctx context.Context, routeId string, instanceId string, add bool)
		// GetInstanceCount
		GetInstanceCount(ctx context.Context) int
	}
)
