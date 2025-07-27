// SPDX-FileCopyrightText: 2025 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

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
