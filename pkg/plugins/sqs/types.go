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

package sqs

import (
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/rs/zerolog"
	"github.com/xmidt-org/ears/pkg/event"
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
	Name     = "sqs"
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
	QueueUrl:            "",
	MaxNumberOfMessages: pointer.Int(10),
	VisibilityTimeout:   pointer.Int(10),
	WaitTimeSeconds:     pointer.Int(10),
	AcknowledgeTimeout:  pointer.Int(5),
	NumRetries:          pointer.Int(0),
	ReceiverQueueDepth:  pointer.Int(100),
	ReceiverPoolSize:    pointer.Int(1),
	NeverDelete:         pointer.Bool(false),
}

type ReceiverConfig struct {
	QueueUrl            string `json:"queueUrl,omitempty"`
	MaxNumberOfMessages *int   `json:"maxNumberOfMessages,omitempty"`
	VisibilityTimeout   *int   `json:"visibilityTimeout,omitempty"`
	WaitTimeSeconds     *int   `json:"waitTimeSeconds,omitempty"`
	AcknowledgeTimeout  *int   `json:"acknowledgeTimeout,omitempty"`
	NumRetries          *int   `json:"numRetries,omitempty"`
	ReceiverQueueDepth  *int   `json:"receiverQueueDepth,omitempty"`
	ReceiverPoolSize    *int   `json:"receiverPoolSize,omitempty"`
	NeverDelete         *bool  `json:"neverDelete,omitempty"`
}

type Receiver struct {
	sync.Mutex
	done                chan struct{}
	stopped             bool
	config              ReceiverConfig
	name                string
	plugin              string
	tid                 tenant.Id
	next                receiver.NextFn
	logger              *zerolog.Logger
	receiveCount        int
	deleteCount         int
	startTime           time.Time
	eventSuccessCounter metric.BoundInt64Counter
	eventFailureCounter metric.BoundInt64Counter
	eventBytesCounter   metric.BoundInt64Counter
}

var DefaultSenderConfig = SenderConfig{
	QueueUrl:            "",
	MaxNumberOfMessages: pointer.Int(10),
	SendTimeout:         pointer.Int(1),
	DelaySeconds:        pointer.Int(0),
}

// SenderConfig can be passed into NewSender() in order to configure
// the behavior of the sender.
type SenderConfig struct {
	QueueUrl            string `json:"queueUrl,omitempty"`
	MaxNumberOfMessages *int   `json:"maxNumberOfMessages,omitempty"`
	SendTimeout         *int   `json:"sendTimeout,omitempty"`
	DelaySeconds        *int   `json:"delaySeconds,omitempty"`
}

type Sender struct {
	sync.Mutex
	sqsService *sqs.SQS
	name       string
	plugin     string
	tid        tenant.Id
	config     SenderConfig
	count      int
	logger     *zerolog.Logger
	eventBatch []event.Event
	done       chan struct{}
}
