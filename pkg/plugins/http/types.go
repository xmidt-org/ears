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

package http

import (
	"github.com/rs/zerolog"
	"github.com/xmidt-org/ears/pkg/errs"
	pkgplugin "github.com/xmidt-org/ears/pkg/plugin"
	"github.com/xmidt-org/ears/pkg/receiver"
	"github.com/xmidt-org/ears/pkg/sender"
	"github.com/xmidt-org/ears/pkg/tenant"
	"github.com/xorcare/pointer"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/propagation"
	"net/http"
	"sync"
)

var _ sender.Sender = (*Sender)(nil)
var _ receiver.Receiver = (*Receiver)(nil)

var (
	Name     = "http"
	Version  = "v0.0.0"
	CommitID = ""
)

func NewPlugin() (*pkgplugin.Plugin, error) {
	return NewPluginVersion(Name, Version, CommitID)
}

func NewPluginVersion(name string, version string, commitID string) (*pkgplugin.Plugin, error) {
	return pkgplugin.NewPlugin(
		pkgplugin.WithName(name),
		pkgplugin.WithVersion(version),
		pkgplugin.WithCommitID(commitID),
		pkgplugin.WithNewReceiver(NewReceiver),
		pkgplugin.WithNewSender(NewSender),
	)
}

type Status struct {
	Code    int    `json:"code,omitempty" xml:"code,omitempty"`
	Message string `json:"message,omitempty" xml:"message,omitempty"`
}

type Tracing struct {
	TraceId string `json:"traceId,omitempty" xml:"traceId,omitempty"`
	//SpanId  string `json:"spanId,omitempty" xml:"spanId,omitempty"`
}

type Response struct {
	Status  *Status  `json:"status,omitempty" xml:"status,omitempty"`
	Tracing *Tracing `json:"tx,omitempty" xml:"tx,omitempty"`
	//Item  interface{} `json:"item,omitempty" xml:"item,omitempty"`
	//Items interface{} `json:"items,omitempty" xml:"items,omitempty"`
	//Data  interface{} `json:"data,omitempty" xml:"data,omitempty"`
}

type ReceiverConfig struct {
	Path               string `json:"path"`
	Method             string `json:"method"`
	Port               *int   `json:"port"`
	TracePayloadOnNack *bool  `json:"tracePayloadOnNack,omitempty"`
	SuccessStatus      *int   `json:"successStatus"`
	FailureStatus      *int   `json:"failureStatus"`
}

var DefaultReceiverConfig = ReceiverConfig{
	SuccessStatus:      pointer.Int(200),
	FailureStatus:      pointer.Int(400),
	TracePayloadOnNack: pointer.Bool(false),
}

type Receiver struct {
	sync.Mutex
	logger              *zerolog.Logger
	srv                 *http.Server
	config              ReceiverConfig
	name                string
	plugin              string
	tid                 tenant.Id
	eventSuccessCounter metric.BoundInt64Counter
	eventFailureCounter metric.BoundInt64Counter
	eventBytesCounter   metric.BoundInt64Counter
	next                receiver.NextFn
}

type SenderConfig struct {
	Url    string `json:"url"`
	Method string `json:"method"`
}

type Sender struct {
	client              *http.Client
	config              SenderConfig
	name                string
	plugin              string
	tid                 tenant.Id
	eventSuccessCounter metric.BoundInt64Counter
	eventFailureCounter metric.BoundInt64Counter
	eventBytesCounter   metric.BoundInt64Counter
	eventProcessingTime metric.BoundInt64Histogram
	eventSendOutTime    metric.BoundInt64Histogram
	b3Propagator        propagation.TextMapPropagator
}

type BadHttpStatusError struct {
	statusCode int
}

func (e *BadHttpStatusError) Error() string {
	return errs.String("BadHttpStatusError", map[string]interface{}{"statusCode": e.statusCode}, nil)
}
