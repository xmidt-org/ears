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

package gears

import (
	"github.com/Shopify/sarama"
	"github.com/rs/zerolog"
	pkgplugin "github.com/xmidt-org/ears/pkg/plugin"
	"github.com/xmidt-org/ears/pkg/secret"
	"github.com/xmidt-org/ears/pkg/sender"
	"github.com/xmidt-org/ears/pkg/tenant"
	"github.com/xorcare/pointer"
	"go.opentelemetry.io/otel/metric"
)

var _ sender.Sender = (*Sender)(nil)

var (
	Name     = "gears"
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
		pkgplugin.WithNewSender(NewSender),
	)
}

var DefaultSenderConfig = SenderConfig{
	Brokers:             "localhost:9092",
	Topic:               "",
	ChannelBufferSize:   pointer.Int(0),
	Username:            "",
	Password:            "",
	CACert:              "",
	AccessCert:          "",
	AccessKey:           "",
	Version:             "",
	SenderPoolSize:      pointer.Int(1),
	DynamicMetricLabels: make([]DynamicMetricLabel, 0),
	Location:            "",
	App:                 "",
	Partner:             "",
	Uses:                "",
	Enveloped:           false,
}

// SenderConfig can be passed into NewSender() in order to configure
// the behavior of the sender.
type SenderConfig struct {
	Clusters            map[string]BrokerSettings `json:"clusters,omitempty"`
	ActiveClusters      string                    `json:"activeClusters,omitempty"`
	Topic               string                    `json:"topic,omitempty"`
	Version             string                    `json:"version,omitempty"`
	ChannelBufferSize   *int                      `json:"channelBufferSize,omitempty"`
	SenderPoolSize      *int                      `json:"senderPoolSize,omitempty"`
	CompressionMethod   string                    `json:"compressionMethod,omitempty"`
	CompressionLevel    *int                      `json:"compressionLevel,omitempty"`
	DynamicMetricLabels []DynamicMetricLabel      `json:"dynamicMetricLabel,omitempty"`
	Location            interface{}               `json:"location,omitempty"`  // gears config: string or array of strings, may contain path
	App                 string                    `json:"app,omitempty"`       // gears config
	Partner             string                    `json:"partner,omitempty"`   // gears config
	Uses                string                    `json:"uses,omitempty"`      // gears config
	Enveloped           bool                      `json:"enveloped,omitempty"` // gears config

	Brokers    string `json:"brokers,omitempty"`    //deprecating. Please specify in cluster
	Username   string `json:"username,omitempty"`   //deprecating. Please specify in cluster
	Password   string `json:"password,omitempty"`   //deprecating. Please specify in cluster
	CACert     string `json:"caCert,omitempty"`     //deprecating. Please specify in cluster
	AccessCert string `json:"accessCert,omitempty"` //deprecating. Please specify in cluster
	AccessKey  string `json:"accessKey,omitempty"`  //deprecating. Please specify in cluster
	TLSEnable  bool   `json:"tlsEnable,omitempty"`  //deprecating. Please specify in cluster
}

type BrokerSettings struct {
	Brokers    string `json:"brokers,omitempty"`
	Username   string `json:"username,omitempty"`
	Password   string `json:"password,omitempty"`
	CACert     string `json:"caCert,omitempty"`
	AccessCert string `json:"accessCert,omitempty"`
	AccessKey  string `json:"accessKey,omitempty"`
	TLSEnable  bool   `json:"tlsEnable,omitempty"`
}

type DynamicMetricLabel struct {
	Label string `json:"label,omitempty"`
	Path  string `json:"path,omitempty"`
}

type DynamicMetricValue struct {
	Label string `json:"label,omitempty"`
	Value string `json:"value,omitempty"`
}

type SenderMetrics struct {
	eventSuccessCounter metric.BoundInt64Counter
	eventFailureCounter metric.BoundInt64Counter
	eventBytesCounter   metric.BoundInt64Counter
	eventProcessingTime metric.BoundInt64Histogram
	eventSendOutTime    metric.BoundInt64Histogram
}

type Sender struct {
	pkgplugin.MetricPlugin
	name      string
	plugin    string
	tid       tenant.Id
	config    SenderConfig
	count     int
	logger    *zerolog.Logger
	producers []*Producer
	stopped   bool
	secrets   secret.Vault
	metrics   map[string]*SenderMetrics
}

type ManualHashPartitioner struct {
	sarama.Partitioner
}

type Producer struct {
	pool   chan sarama.SyncProducer
	done   chan bool
	client sarama.Client
	sender *Sender
	logger *zerolog.Logger
	key    string
}
