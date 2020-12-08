// Copyright 2020 Comcast Cable Communications Management, LLC
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

package handlers

import (
	"encoding/json"
	"net/http"
)

// Config interface for uber/fx
type Config interface {
	GetString(key string) string
	GetInt(key string) int
	GetBool(key string) bool
}

// Logging expected keys
const (
	LogTraceId      = "tx.traceId"
	LogSpanId       = "tx.spanId"
	LogParentSpanId = "tx.parentSpanId"
	LogChildSpanId  = "tx.childSpanId"
	LogTenantId     = "gears.app.id"
)

// API header expected keys
const (
	HeaderTraceId      = "X-B3-TraceId"
	HeaderSpanId       = "X-B3-SpanId"
	HeaderParentSpanId = "X-B3-ParentSpanId"
	HeaderTenantId     = "Application-Id"
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

func (r Response) Respond(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	if r.Status != nil {
		r.Status.Message = http.StatusText(r.Status.Code)
		w.WriteHeader(r.Status.Code)
	}
	json.NewEncoder(w).Encode(r)
}
