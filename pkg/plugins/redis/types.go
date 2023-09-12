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

package redis

import (
	"github.com/go-redis/redis"
	"github.com/rs/zerolog"
	"github.com/xmidt-org/ears/pkg/tenant"
	"github.com/xorcare/pointer"
	"go.opentelemetry.io/otel/metric"
	"time"

	pkgplugin "github.com/xmidt-org/ears/pkg/plugin"

	"github.com/xmidt-org/ears/pkg/receiver"
	"github.com/xmidt-org/ears/pkg/sender"
)

var _ sender.Sender = (*Sender)(nil)
var _ receiver.Receiver = (*Receiver)(nil)

var (
	Name     = "redis"
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

var DefaultReceiverConfig = ReceiverConfig{
	Endpoint:           "localhost:6379",
	Channel:            "ears",
	TracePayloadOnNack: pointer.Bool(false),
}

type ReceiverConfig struct {
	Endpoint           string `json:"endpoint,omitempty"`
	Channel            string `json:"channel,omitempty"`
	TracePayloadOnNack *bool  `json:"tracePayloadOnNack,omitempty"`
}

type Receiver struct {
	pkgplugin.MetricPlugin
	stopped             bool
	redisClient         *redis.Client
	pubsub              *redis.PubSub
	done                chan struct{}
	config              ReceiverConfig
	name                string
	plugin              string
	tid                 tenant.Id
	next                receiver.NextFn
	logger              *zerolog.Logger
	startTime           time.Time
	eventSuccessCounter metric.BoundInt64Counter
	eventFailureCounter metric.BoundInt64Counter
	eventBytesCounter   metric.BoundInt64Counter
}

var DefaultSenderConfig = SenderConfig{
	Endpoint: "localhost:6379",
	Channel:  "ears",
}

// SenderConfig can be passed into NewSender() in order to configure
// the behavior of the sender.
type SenderConfig struct {
	Endpoint string `json:"endpoint,omitempty"`
	Channel  string `json:"channel,omitempty"`
}

type Sender struct {
	pkgplugin.MetricPlugin
	name                string
	plugin              string
	tid                 tenant.Id
	config              SenderConfig
	logger              *zerolog.Logger
	client              *redis.Client
	eventSuccessCounter metric.BoundInt64Counter
	eventFailureCounter metric.BoundInt64Counter
	eventBytesCounter   metric.BoundInt64Counter
	eventProcessingTime metric.BoundInt64Histogram
	eventSendOutTime    metric.BoundInt64Histogram
}
