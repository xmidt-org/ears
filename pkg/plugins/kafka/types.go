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

package kafka

import (
	"context"
	"github.com/Shopify/sarama"
	"github.com/rs/zerolog"
	"github.com/xmidt-org/ears/pkg/secret"
	"github.com/xmidt-org/ears/pkg/tenant"
	"github.com/xorcare/pointer"
	"go.opentelemetry.io/otel/metric"
	"sync"
	"time"

	pkgplugin "github.com/xmidt-org/ears/pkg/plugin"

	"github.com/xmidt-org/ears/pkg/receiver"
	"github.com/xmidt-org/ears/pkg/sender"
)

var _ sender.Sender = (*Sender)(nil)
var _ receiver.Receiver = (*Receiver)(nil)

var (
	Name     = "kafka"
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
	Brokers:           "localhost:9092",
	Topic:             "quickstart-events",
	GroupId:           "",
	Username:          "",
	Password:          "",
	CACert:            "",
	AccessCert:        "",
	AccessKey:         "",
	Version:           "",
	CommitInterval:    pointer.Int(1),
	ChannelBufferSize: pointer.Int(0),
}

type ReceiverConfig struct {
	Brokers             string `json:"brokers,omitempty"`
	Topic               string `json:"topic,omitempty"`
	GroupId             string `json:"groupId,omitempty"`
	Username            string `json:"username,omitempty"` // yaml
	Password            string `json:"password,omitempty"`
	CACert              string `json:"caCert,omitempty"`
	AccessCert          string `json:"accessCert,omitempty"`
	AccessKey           string `json:"accessKey,omitempty"`
	Version             string `json:"version,omitempty"`
	CommitInterval      *int   `json:"commitInterval,omitempty"`
	ChannelBufferSize   *int   `json:"channelBufferSize,omitempty"`
	ConsumeByPartitions bool   `json:"consumeByPartitions,omitempty"`
	TLSEnable           bool   `json:"tlsEnable,omitempty"`
}

type Receiver struct {
	sync.Mutex
	done      chan struct{}
	stopped   bool
	config    ReceiverConfig
	name      string
	plugin    string
	tid       tenant.Id
	next      receiver.NextFn
	logger    *zerolog.Logger
	count     int
	startTime time.Time
	sarama.ConsumerGroupSession
	wg                  sync.WaitGroup
	ready               chan bool
	cancel              context.CancelFunc
	ctx                 context.Context
	client              sarama.ConsumerGroup
	topics              []string
	handler             func(message *sarama.ConsumerMessage) bool
	eventSuccessCounter metric.BoundInt64Counter
	eventFailureCounter metric.BoundInt64Counter
	eventBytesCounter   metric.BoundInt64Counter
	secrets             secret.Vault
}

var DefaultSenderConfig = SenderConfig{
	Brokers:           "localhost:9092",
	Topic:             "quickstart-events",
	Partition:         pointer.Int(-1),
	ChannelBufferSize: pointer.Int(0),
	Username:          "",
	Password:          "",
	CACert:            "",
	AccessCert:        "",
	AccessKey:         "",
	Version:           "",
	SenderPoolSize:    pointer.Int(1),
	PartitionPath:     "",
}

// SenderConfig can be passed into NewSender() in order to configure
// the behavior of the sender.
type SenderConfig struct {
	Brokers           string `json:"brokers,omitempty"`
	Topic             string `json:"topic,omitempty"`
	Partition         *int   `json:"partition,omitempty"`
	PartitionPath     string `json:"partitionPath,omitempty"` // if path is set look up partition from event rather than using the hard coded partition id
	Username          string `json:"username,omitempty"`
	Password          string `json:"password,omitempty"`
	CACert            string `json:"caCert,omitempty"`
	AccessCert        string `json:"accessCert,omitempty"`
	AccessKey         string `json:"accessKey,omitempty"`
	Version           string `json:"version,omitempty"`
	ChannelBufferSize *int   `json:"channelBufferSize,omitempty"`
	TLSEnable         bool   `json:"tlsEnable,omitempty"`
	SenderPoolSize    *int   `json:"senderPoolSize,omitempty"`
}

type Sender struct {
	sync.Mutex
	name                string
	plugin              string
	tid                 tenant.Id
	config              SenderConfig
	count               int
	logger              *zerolog.Logger
	producer            *Producer
	stopped             bool
	eventSuccessCounter metric.BoundInt64Counter
	eventFailureCounter metric.BoundInt64Counter
	eventBytesCounter   metric.BoundInt64Counter
	eventProcessingTime metric.BoundInt64ValueRecorder
	secrets             secret.Vault
}

type ManualHashPartitioner struct {
	sarama.Partitioner
}

type Producer struct {
	pool   chan sarama.SyncProducer
	done   chan bool
	client sarama.Client
	logger *zerolog.Logger
}
