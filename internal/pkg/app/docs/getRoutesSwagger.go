// SPDX-FileCopyrightText: 2025 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

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

type RouteConfig struct {
	route.Config
}
