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

// swagger:route GET /v1/receivers admin getAllReceivers
// Gets list of all receiver plugin instances currently present along with their reference count.
// responses:
//   200: ReceiversResponse
//   500: ReceiversErrorResponse

import (
	"github.com/xmidt-org/ears/internal/pkg/plugin"
)

// Items response containing a list of receiver plugin instances.
// swagger:response receiversResponse
type receiversResponseWrapper struct {
	// in: body
	Body ReceiversResponse
}

// Item response containing a receivers error.
// swagger:response receiversErrorResponse
type receiversErrorResponseWrapper struct {
	// in: body
	Body ReceiversErrorResponse
}

type ReceiversResponse struct {
	Status responseStatus                   `json:"status"`
	Items  map[string]plugin.ReceiverStatus `json:"items"`
}

type ReceiversErrorResponse struct {
	Status responseStatus `json:"status"`
	Item   string         `json:"item"`
}
