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

package nop

import (
	"github.com/rs/zerolog"
	"github.com/xmidt-org/ears/pkg/tenant"
	"go.opentelemetry.io/otel/metric"
	"sync"

	pkgplugin "github.com/xmidt-org/ears/pkg/plugin"

	"github.com/xmidt-org/ears/pkg/receiver"
	"github.com/xmidt-org/ears/pkg/sender"
)

var _ sender.Sender = (*Sender)(nil)
var _ receiver.Receiver = (*Receiver)(nil)

var (
	Name     = "nop"
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

var DefaultReceiverConfig = ReceiverConfig{}

type ReceiverConfig struct {
}

type Receiver struct {
	sync.Mutex
	done                          chan struct{}
	stopped                       bool
	config                        ReceiverConfig
	name                          string
	plugin                        string
	tid                           tenant.Id
	next                          receiver.NextFn
	logger                        zerolog.Logger
	successCounter                int
	errorCounter                  int
	successVelocityCounter        int
	errorVelocityCounter          int
	currentSuccessVelocityCounter int
	currentErrorVelocityCounter   int
	currentSec                    int
}

var DefaultSenderConfig = SenderConfig{}

type SenderConfig struct {
}

type Sender struct {
	sync.Mutex
	name                string
	plugin              string
	tid                 tenant.Id
	config              SenderConfig
	eventSuccessCounter metric.BoundInt64Counter
	eventFailureCounter metric.BoundInt64Counter
	eventBytesCounter   metric.BoundInt64Counter
	eventProcessingTime metric.BoundInt64Histogram
	eventSendOutTime    metric.BoundInt64Histogram
}
