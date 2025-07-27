// SPDX-FileCopyrightText: 2025 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package docs

// swagger:route GET /v1/filters admin getAllFilters
// Gets list of all filter plugin instances currently present along with their reference count.
// responses:
//   200: FiltersResponse
//   500: FiltersErrorResponse

// swagger:route GET /v1/orgs/{orgId}/applications/{appId}/filters plugins getFilters
// Gets list of all filters currently present for a single tenant.
// responses:
//   200: FiltersResponse
//   500: FiltersErrorResponse

import (
	"github.com/xmidt-org/ears/internal/pkg/plugin"
)

// Items response containing a list of filter plugin instances.
// swagger:response receiversResponse
type filtersResponseWrapper struct {
	// in: body
	Body FiltersResponse
}

// Item response containing a filters error.
// swagger:response filtersErrorResponse
type filtersErrorResponseWrapper struct {
	// in: body
	Body FiltersErrorResponse
}

type FiltersResponse struct {
	Status responseStatus                 `json:"status"`
	Items  map[string]plugin.FilterStatus `json:"items"`
}

type FiltersErrorResponse struct {
	Status responseStatus `json:"status"`
	Item   string         `json:"item"`
}
