// SPDX-FileCopyrightText: 2025 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package docs

// swagger:route GET /v1/tenants admin getAllTenants
// Gets list of all tenant configs including their event quota.
// responses:
//   200: TenantsResponse
//   500: TenantErrorResponse

// Item response containing list of tenants.
// swagger:response tenantsResponse
type tenantsResponseWrapper struct {
	// in: body
	Body TenantsResponse
}

type TenantsResponse struct {
	Status responseStatus `json:"status"`
	Item   []TenantConfig `json:"item"`
}
