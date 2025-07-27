// SPDX-FileCopyrightText: 2025 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package docs

// swagger:route DELETE /v1/orgs/{orgId}/applications/{appId}/config tenants deleteTenant
// Removes an existing tenant from the system provided tenant has no routes.
// responses:
//   200: TenantDeleteResponse
//   500: TenantErrorResponse

// Item response.
// swagger:response tenantDeleteResponse
type tenantDeleteResponseWrapper struct {
	// in: body
	Body TenantDeleteResponse
}

type TenantDeleteResponse struct {
	Status responseStatus `json:"status"`
	Item   string         `json:"item"`
}
