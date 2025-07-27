// SPDX-FileCopyrightText: 2025 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/goccy/go-yaml"
	"github.com/xmidt-org/ears/internal/pkg/rtsemconv"
)

// API header expected keys
const (
	HeaderTraceId  = "X-B3-TraceId"
	HeaderTenantId = "Application-Id"
)

// #######################################################
// API Response
// #######################################################

type Status struct {
	Code    int    `json:"code,omitempty" xml:"code,omitempty"`
	Message string `json:"message,omitempty" xml:"message,omitempty"`
}

type Tracing struct {
	TraceId string `json:"traceId,omitempty" xml:"traceId,omitempty"`
	SpanId  string `json:"spanId,omitempty" xml:"spanId,omitempty"`
}

type Response struct {
	Status  *Status  `json:"status,omitempty" xml:"status,omitempty"`
	Tracing *Tracing `json:"tx,omitempty" xml:"tx,omitempty"`

	Item  interface{} `json:"item,omitempty" xml:"item,omitempty"`
	Items interface{} `json:"items,omitempty" xml:"items,omitempty"`
	Data  interface{} `json:"data,omitempty" xml:"data,omitempty"`
}

func (r Response) Respond(ctx context.Context, w http.ResponseWriter, doYaml bool) {
	if doYaml {
		w.Header().Set("Content-Type", "text/x-yaml")
	} else {
		w.Header().Set("Content-Type", "application/json")
	}
	traceId, ok := ctx.Value(rtsemconv.EarsLogTraceIdKey).(string)
	if ok && traceId != "" {
		r.Tracing.TraceId = traceId
	}
	if r.Status != nil {
		r.Status.Message = http.StatusText(r.Status.Code)
		w.WriteHeader(r.Status.Code)
	}
	if doYaml {
		yaml.NewEncoder(w).Encode(r)
	} else {
		json.NewEncoder(w).Encode(r)
	}
}

func ErrorResponse(apiErr ApiError) Response {
	return Response{
		Status: &Status{
			Code: apiErr.StatusCode(),
		},
		Item: apiErr.Error(),
	}
}

func ItemResponse(item interface{}) Response {
	return Response{
		Status: &Status{
			Code: http.StatusOK,
		},
		Item: item,
	}
}

func ItemStatusResponse(item interface{}, status int) Response {
	return Response{
		Status: &Status{
			Code: status,
		},
		Item: item,
	}
}

func ItemsResponse(item interface{}) Response {
	return Response{
		Status: &Status{
			Code: http.StatusOK,
		},
		Items: item,
	}
}

func SimpleResponse(ctx context.Context) Response {
	return Response{
		Status: &Status{
			Code: http.StatusOK,
		},
	}
}
