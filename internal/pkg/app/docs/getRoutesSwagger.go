// Copyright 2021 Comcast Cable Communications Management, LLC
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

package docs

// swagger:route GET /v1/orgs/{orgId}/applications/{appId}/routes routes getRoutes
// Gets list of all routes currently present in the routing table for a single tenant.
// responses:
//   200: RoutesResponse
//   500: RouteErrorResponse

import "github.com/xmidt-org/ears/pkg/route"

// Items response containing a list of routes.
// swagger:response routesResponse
type routesResponseWrapper struct {
	// in: body
	Body RoutesResponse
}

type RoutesResponse struct {
	Status responseStatus `json:"status"`
	Items  []RouteConfig  `json:"items"`
}

// swagger:model RouteConfig
type RouteConfig route.Config
