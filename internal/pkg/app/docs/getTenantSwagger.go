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
