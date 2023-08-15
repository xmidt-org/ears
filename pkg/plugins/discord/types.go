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

package discord

import (
	"sync"

	"github.com/bwmarrin/discordgo"
	"github.com/rs/zerolog"
	pkgplugin "github.com/xmidt-org/ears/pkg/plugin"
	"github.com/xmidt-org/ears/pkg/receiver"
	"github.com/xmidt-org/ears/pkg/sender"
	"github.com/xmidt-org/ears/pkg/sharder"
	"github.com/xmidt-org/ears/pkg/tenant"
	"github.com/xorcare/pointer"
	"go.opentelemetry.io/otel/metric"
)

var _ receiver.Receiver = (*Receiver)(nil)
var _ sender.Sender = (*Sender)(nil)

var (
	Name     = "discord"
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

type ReceiverConfig struct {
	BotToken        string `json:"botToken"`
	UseShardMonitor *bool  `json:"useShardMonitor,omitempty"`
}

type SenderConfig struct {
	BotToken  string `json:"botToken"`
	ChannelId string `json:"channelId"`
}

var DefaultReceiverConfig = ReceiverConfig{
	BotToken:        "",
	UseShardMonitor: pointer.Bool(false),
}

type Sender struct {
	sync.Mutex
	sess                          *discordgo.Session
	config                        SenderConfig
	name                          string
	plugin                        string
	tid                           tenant.Id
	eventSuccessCounter           metric.BoundInt64Counter
	eventFailureCounter           metric.BoundInt64Counter
	eventBytesCounter             metric.BoundInt64Counter
	eventProcessingTime           metric.BoundInt64Histogram
	eventSendOutTime              metric.BoundInt64Histogram
	successCounter                int
	errorCounter                  int
	successVelocityCounter        int
	errorVelocityCounter          int
	currentSuccessVelocityCounter int
	currentErrorVelocityCounter   int
	currentSec                    int64
}

type Receiver struct {
	sync.Mutex
	stopped                        bool
	shardDistributor               sharder.ShardDistributor
	shardMonitorStopChannel        chan bool
	shardUpdateListenerStopChannel chan bool
	shardConfig                    sharder.ShardConfig
	stopChannelMap                 map[int]chan bool
	sess                           *discordgo.Session
	shardsCount                    *int
	logger                         *zerolog.Logger
	config                         ReceiverConfig
	name                           string
	plugin                         string
	tid                            tenant.Id
	event                          *discordgo.Message
	eventSuccessCounter            metric.BoundInt64Counter
	eventFailureCounter            metric.BoundInt64Counter
	eventBytesCounter              metric.BoundInt64Counter
	next                           receiver.NextFn
	successCounter                 int
	errorCounter                   int
	successVelocityCounter         int
	errorVelocityCounter           int
	currentSuccessVelocityCounter  int
	currentErrorVelocityCounter    int
	currentSec                     int64
}
