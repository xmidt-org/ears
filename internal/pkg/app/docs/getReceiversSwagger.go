// SPDX-FileCopyrightText: 2025 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package docs

// swagger:route GET /v1/receivers admin getAllReceivers
// Gets list of all receiver plugin instances currently present along with their reference count.
// responses:
//   200: ReceiversResponse
//   500: ReceiversErrorResponse

// swagger:route GET /v1/orgs/{orgId}/applications/{appId}/receivers plugins getReceivers
// Gets list of all receivers currently present for a single tenant.
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
