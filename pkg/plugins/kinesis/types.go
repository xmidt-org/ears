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

package kinesis

import (
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/service/kinesis"
	"github.com/rs/zerolog"
	"github.com/xmidt-org/ears/internal/pkg/syncer"
	"github.com/xmidt-org/ears/pkg/errs"
	"github.com/xmidt-org/ears/pkg/event"
	pkgplugin "github.com/xmidt-org/ears/pkg/plugin"
	"github.com/xmidt-org/ears/pkg/secret"
	"github.com/xmidt-org/ears/pkg/sharder"
	"github.com/xmidt-org/ears/pkg/tenant"
	"github.com/xorcare/pointer"
	"go.opentelemetry.io/otel/metric"
	"sync"

	"github.com/xmidt-org/ears/pkg/receiver"
	"github.com/xmidt-org/ears/pkg/sender"
)

var _ sender.Sender = (*Sender)(nil)
var _ receiver.Receiver = (*Receiver)(nil)

var (
	Name     = "kinesis"
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
	StreamName:              "",
	AcknowledgeTimeout:      pointer.Int(5),
	ShardIteratorType:       kinesis.ShardIteratorTypeLatest, // AT_SEQUENCE_NUMBER, AFTER_SEQUENCE_NUMBER, LATEST, AT_TIMESTAMP, TRIM_HORIZON
	TracePayloadOnNack:      pointer.Bool(false),
	EnhancedFanOut:          pointer.Bool(false),
	ConsumerName:            "",
	AWSRoleARN:              "",
	AWSSecretAccessKey:      "",
	AWSAccessKeyId:          "",
	AWSRegion:               endpoints.UsWest2RegionID,
	MaxCheckpointAgeSeconds: pointer.Int(86400),
	UseCheckpoint:           pointer.Bool(true),
	UseShardMonitor:         pointer.Bool(false),
	StartingTimestamp:       pointer.Int64(0),
	StartingSequenceNumber:  "",
	EmptyStreamWaitSeconds:  pointer.Int(1),
}

type ReceiverConfig struct {
	StreamName              string `json:"streamName,omitempty"`
	AcknowledgeTimeout      *int   `json:"acknowledgeTimeout,omitempty"`
	ShardIteratorType       string `json:"shardIteratorType,omitempty"`
	TracePayloadOnNack      *bool  `json:"tracePayloadOnNack,omitempty"`
	EnhancedFanOut          *bool  `json:"enhancedFanOut,omitempty"`
	ConsumerName            string `json:"consumerName,omitempty"` // enhanced fan-out only
	AWSRoleARN              string `json:"awsRoleARN,omitempty"`
	AWSAccessKeyId          string `json:"awsAccessKeyId,omitempty"`
	AWSSecretAccessKey      string `json:"awsSecretAccessKey,omitempty"`
	AWSRegion               string `json:"awsRegion,omitempty"`
	UseCheckpoint           *bool  `json:"useCheckpoint,omitempty"`
	MaxCheckpointAgeSeconds *int   `json:"maxCheckpointAgeSeconds,omitempty"`
	UseShardMonitor         *bool  `json:"useShardMonitor,omitempty"`
	StartingSequenceNumber  string `json:"startingSequenceNumber,omitempty"`
	StartingTimestamp       *int64 `json:"startingTimestamp,omitempty"`
	EmptyStreamWaitSeconds  *int   `json:"emptyStreamWaitSeconds,omitempty"`
}

type Receiver struct {
	sync.Mutex
	done                           chan struct{}
	stopped                        bool
	stopChannelMap                 map[int]chan bool
	shardMonitorStopChannel        chan bool
	shardUpdateListenerStopChannel chan bool
	shardDistributor               sharder.ShardDistributor
	config                         ReceiverConfig
	name                           string
	plugin                         string
	tid                            tenant.Id
	next                           receiver.NextFn
	logger                         *zerolog.Logger
	shardConfig                    sharder.ShardConfig
	svc                            *kinesis.Kinesis
	consumer                       *kinesis.DescribeStreamConsumerOutput
	stream                         *kinesis.DescribeStreamOutput
	secrets                        secret.Vault
	eventSuccessCounter            metric.BoundInt64Counter
	eventFailureCounter            metric.BoundInt64Counter
	eventBytesCounter              metric.BoundInt64Counter
	eventLagMillis                 metric.BoundInt64Histogram
	eventTrueLagMillis             metric.BoundInt64Histogram
	awsRoleArn                     string
	awsAccessKey                   string
	awsAccessSecret                string
	awsRegion                      string
	streamName                     string
	consumerName                   string
	successCounter                 int
	errorCounter                   int
	successVelocityCounter         int
	errorVelocityCounter           int
	currentSuccessVelocityCounter  int
	currentErrorVelocityCounter    int
	currentSec                     int64
	tableSyncer                    syncer.DeltaSyncer
}

var DefaultSenderConfig = SenderConfig{
	StreamName:          "",
	MaxNumberOfMessages: pointer.Int(1),
	SendTimeout:         pointer.Int(1),
	PartitionKey:        "",
	PartitionKeyPath:    "",
	AWSRoleARN:          "",
	AWSSecretAccessKey:  "",
	AWSAccessKeyId:      "",
	AWSRegion:           endpoints.UsWest2RegionID,
}

type SenderConfig struct {
	StreamName          string `json:"streamName,omitempty"`
	MaxNumberOfMessages *int   `json:"maxNumberOfMessages,omitempty"`
	SendTimeout         *int   `json:"sendTimeout,omitempty"`
	PartitionKey        string `json:"partitionKey,omitempty"`
	PartitionKeyPath    string `json:"partitionKeyPath,omitempty"`
	AWSRoleARN          string `json:"awsRoleARN,omitempty"`
	AWSAccessKeyId      string `json:"awsAccessKeyId,omitempty"`
	AWSSecretAccessKey  string `json:"awsSecretAccessKey,omitempty"`
	AWSRegion           string `json:"awsRegion,omitempty"`
}

type Sender struct {
	sync.Mutex
	kinesisService                *kinesis.Kinesis
	name                          string
	plugin                        string
	tid                           tenant.Id
	config                        SenderConfig
	logger                        *zerolog.Logger
	eventBatch                    []event.Event
	done                          chan struct{}
	secrets                       secret.Vault
	eventSuccessCounter           metric.BoundInt64Counter
	eventFailureCounter           metric.BoundInt64Counter
	eventBytesCounter             metric.BoundInt64Counter
	eventProcessingTime           metric.BoundInt64Histogram
	eventSendOutTime              metric.BoundInt64Histogram
	awsRoleArn                    string
	awsAccessKey                  string
	awsAccessSecret               string
	awsRegion                     string
	streamName                    string
	successCounter                int
	errorCounter                  int
	successVelocityCounter        int
	errorVelocityCounter          int
	currentSuccessVelocityCounter int
	currentErrorVelocityCounter   int
	currentSec                    int64
	tableSyncer                   syncer.DeltaSyncer
}

type KinesisError struct {
	op  string
	err error
}

func (e *KinesisError) Error() string {
	return errs.String("KinesisError", map[string]interface{}{"op": e.op}, e.err)
}

func (e *KinesisError) Unwrap() error {
	return e.err
}
