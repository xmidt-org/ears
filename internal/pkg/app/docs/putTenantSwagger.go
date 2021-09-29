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

// swagger:route PUT /v1/orgs/{orgId}/applications/{appId}/config tenants putTenant
// Adds a new tenant with org ID, app ID and event quota given in the put body.
// responses:
//   200: TenantResponse
//   500: TenantErrorResponse

import (
	"github.com/xmidt-org/ears/pkg/tenant"
)

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
	Body tenant.Config
}

// swagger:parameters putRoute postRoute getRoute deleteRoute putTenant getTenant deleteTenant
type appIdParamWrapper struct {
	// App ID
	// in: path
	// required: true
	AppId string `json:"appId"`
}

// swagger:parameters putRoute postRoute getRoute deleteRoute putTenant getTenant deleteTenant
type orgIdParamWrapper struct {
	// Org ID
	// in: path
	// required: true
	OrgId string `json:"orgId"`
}

type TenantErrorResponse struct {
	Status responseStatus `json:"status"`
	Item   string         `json:"item"`
}
