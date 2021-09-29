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

// swagger:route DELETE /v1/orgs/{orgId}/applications/{appId}/routes/{routeId} routes deleteRoute
// Removes an existing route from the routing table if a route with the given ID exists.
// responses:
//   200: RouteDeleteResponse
//   500: RouteErrorResponse

// Item response containing the ID of the deleted route.
// swagger:response routeDeleteResponse
type routeDeleteResponseWrapper struct {
	// in: body
	Body RouteDeleteResponse
}

type RouteDeleteResponse struct {
	Status responseStatus `json:"status"`
	Item   string         `json:"item"`
}
