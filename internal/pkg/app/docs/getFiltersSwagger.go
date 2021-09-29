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

// swagger:route GET /v1/filters admin getAllFilters
// Gets list of all filter plugin instances currently present along with their reference count.
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
