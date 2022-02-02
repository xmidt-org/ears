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

// swagger:route GET /v1/senders admin getAllSenders
// Gets list of all sender plugin instances currently present along with their reference count.
// responses:
//   200: SendersResponse
//   500: SendersErrorResponse

// swagger:route GET /v1/orgs/{orgId}/applications/{appId}/senders plugins getSenders
// Gets list of all senders currently present for a single tenant.
// responses:
//   200: SendersResponse
//   500: SendersErrorResponse

import (
	"github.com/xmidt-org/ears/internal/pkg/plugin"
)

// Items response containing a list of sender plugin instances.
// swagger:response sendersResponse
type sendersResponseWrapper struct {
	// in: body
	Body SendersResponse
}

// Item response containing a senders error.
// swagger:response sendersErrorResponse
type sendersErrorResponseWrapper struct {
	// in: body
	Body SendersErrorResponse
}

type SendersResponse struct {
	Status responseStatus                 `json:"status"`
	Items  map[string]plugin.SenderStatus `json:"items"`
}

type SendersErrorResponse struct {
	Status responseStatus `json:"status"`
	Item   string         `json:"item"`
}
