// SPDX-FileCopyrightText: 2025 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package docs

// swagger:route GET /v1/orgs/{orgId}/applications/{appId}/config tenants getTenant
// Gets config including event quota of existing tenant.
// responses:
//   200: TenantResponse
//   500: TenantErrorResponse

import (
	"github.com/xmidt-org/ears/pkg/tenant"
)

// Item response containing tenant.
// swagger:response tenantResponse
type tenantResponseWrapper struct {
	// in: body
	Body TenantResponse
}

type TenantResponse struct {
	Status responseStatus `json:"status"`
	Item   TenantConfig   `json:"item"`
}

type TenantConfig struct {
	tenant.Config
}
