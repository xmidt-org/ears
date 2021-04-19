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
	"github.com/xorcare/pointer"
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
	Brokers:        "",
	Topic:          "",
	GroupId:        "",
	Username:       "",
	Password:       "",
	CACert:         "",
	AccessCert:     "",
	AccessKey:      "",
	Version:        "",
	CommitInterval: pointer.Int(1),
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
	ConsumeByPartitions bool
	TLSEnable           bool
}

type Receiver struct {
	sync.Mutex
	done      chan struct{}
	config    ReceiverConfig
	next      receiver.NextFn
	logger    zerolog.Logger
	count     int
	startTime time.Time

	sarama.ConsumerGroupSession
	wg      sync.WaitGroup
	ready   chan bool
	cancel  context.CancelFunc
	ctx     context.Context
	client  sarama.ConsumerGroup
	topics  []string
	handler func(message *sarama.ConsumerMessage) bool
}

var DefaultSenderConfig = SenderConfig{
	Brokers:   "",
	Topic:     "",
	Partition: pointer.Int(-1),
}

// SenderConfig can be passed into NewSender() in order to configure
// the behavior of the sender.
type SenderConfig struct {
	Brokers   string `json:"brokers,omitempty"`
	Topic     string `json:"topic,omitempty"`
	Partition *int   `json:"partition,omitempty"`
}

type Sender struct {
	sync.Mutex
	config SenderConfig
	count  int
	logger zerolog.Logger
}

func (s *Sender) Unwrap() sender.Sender {
	return s
}
