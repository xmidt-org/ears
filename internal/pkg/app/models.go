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

	"github.com/xmidt-org/ears/pkg/route"
)

/*type (
	// A RoutingEntry represents an entry in the EARS routing table
	RoutingTableEntry struct {
	}
)*/

type (

	// A RoutingTableManager supports modifying and querying an EARS routing table
	RoutingTableManager interface {
		// AddRoute adds a route to live routing table and runs it and also stores the route in the persistence layer
		AddRoute(ctx context.Context, route *route.Config) error
		// RemoveRoute removes a route from a live routing table and stops it and also removes the route from the persistence layer
		RemoveRoute(ctx context.Context, routeId string) error
	}
)
