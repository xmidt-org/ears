// SPDX-FileCopyrightText: 2025 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package docs

// swagger:route PUT /v1/orgs/{orgId}/applications/{appId}/config tenants putTenant
// Adds a new tenant with org ID, app ID and event quota given in the put body.
// responses:
//   200: TenantResponse
//   500: TenantErrorResponse

// Item response containing a route error.
// swagger:response tenantErrorResponse
type tenantErrorResponseWrapper struct {
	// in: body
	Body TenantErrorResponse
}

// swagger:parameters putTenant
type tenantParamWrapper struct {
	// Tenant configuration including event quota.
	// in: body
	// required: true
	Body TenantConfig
}

type TenantErrorResponse struct {
	Status responseStatus `json:"status"`
	Item   string         `json:"item"`
}
