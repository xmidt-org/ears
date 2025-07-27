// SPDX-FileCopyrightText: 2025 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package docs

// swagger:route POST /v1/orgs/{orgId}/applications/{appId}/routes routes postRoute
// Adds a new route to the routing table or updates an existing route. Route ID can be given in the body. If it is omitted a hash will be calculated and used instead.
// responses:
//   200: RouteResponse
//   500: RouteErrorResponse
