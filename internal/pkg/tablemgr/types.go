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
	"github.com/xmidt-org/ears/internal/pkg/plugin"
	"github.com/xmidt-org/ears/internal/pkg/syncer"
	"github.com/xmidt-org/ears/pkg/filter"
	"github.com/xmidt-org/ears/pkg/route"
	"github.com/xmidt-org/ears/pkg/tenant"
)

type (

	// A RoutingTableManager supports modifying and querying an EARS routing table
	RoutingTableManager interface {
		RoutingTableGlobalSyncer // routing table manager delegates to routing table global syncer for startup and tear down
		syncer.LocalSyncer       // to sync routing table upon receipt of an update notification for a single route
		// AddRoute adds a route to live routing table and runs it and also stores the route in the persistence layer
		AddRoute(ctx context.Context, route *route.Config) error
		// RemoveRoute removes a route from a live routing table and stops it and also removes the route from the persistence layer
		RemoveRoute(ctx context.Context, tenantId tenant.Id, routeId string) error
		// GetRoute gets a single route by its ID from persistence layer
		GetRoute(ctx context.Context, tenantId tenant.Id, routeId string) (*route.Config, error)
		// GetAllTenantRoutes gets all routes for a tenant from persistence layer
		GetAllTenantRoutes(ctx context.Context, tenantId tenant.Id) ([]route.Config, error)
		// GetAllRoutes gets all routes from persistence layer
		GetAllRoutes(ctx context.Context) ([]route.Config, error)
		// GetAllSenders gets all senders currently present in the system
		GetAllSendersStatus(ctx context.Context) (map[string]plugin.SenderStatus, error)
		// GetAllReceivers gets all receivers currently present in the system
		GetAllReceiversStatus(ctx context.Context) (map[string]plugin.ReceiverStatus, error)
		// GetAllFilters gets all filters currently present in the system
		GetAllFiltersStatus(ctx context.Context) (map[string]filter.Filterer, error)
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
		// GetAllRegisteredRoutes gets all routes that are currently registered and running on ears instance
		GetAllRegisteredRoutes() ([]route.Config, error)
	}
)
