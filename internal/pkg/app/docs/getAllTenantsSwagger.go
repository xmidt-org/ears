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

// swagger:route GET /v1/tenants admin getAllTenants
// Gets list of all tenant configs including their event quota.
// responses:
//   200: TenantsResponse
//   500: TenantErrorResponse

import (
	"github.com/xmidt-org/ears/pkg/tenant"
)

// Item response containing list of tenants.
// swagger:response tenantsResponse
type tenantsResponseWrapper struct {
	// in: body
	Body TenantsResponse
}

type TenantsResponse struct {
	Status responseStatus  `json:"status"`
	Item   []tenant.Config `json:"item"`
}
