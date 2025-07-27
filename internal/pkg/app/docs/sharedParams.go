// SPDX-FileCopyrightText: 2025 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package docs

// swagger:parameters putRoute postRoute getRoute deleteRoute putTenant getTenant deleteTenant postRouteEvent
type appIdParamWrapper struct {
	// App ID
	// in: path
	// required: true
	AppId string `json:"appId"`
}

// swagger:parameters putRoute postRoute getRoute deleteRoute putTenant getTenant deleteTenant postRouteEvent
type orgIdParamWrapper struct {
	// Org ID
	// in: path
	// required: true
	OrgId string `json:"orgId"`
}
