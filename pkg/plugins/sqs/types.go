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
}

type ReceiverConfig struct {
	QueueUrl            string `json:"queueUrl,omitempty"`
	MaxNumberOfMessages *int   `json:"maxNumberOfMessages,omitempty"`
	VisibilityTimeout   *int   `json:"visibilityTimeout,omitempty"`
	WaitTimeSeconds     *int   `json:"waitTimeSeconds,omitempty"`
	AcknowledgeTimeout  *int   `json:"acknowledgeTimeout,omitempty"`
	NumRetries          *int   `json:"numRetries,omitempty"`
	ReceiverQueueDepth  *int   `json:"receiverQueueDepth,omitempty"`
}

type Receiver struct {
	sync.Mutex
	done         chan struct{}
	stop         bool
	config       ReceiverConfig
	next         receiver.NextFn
	receiveCount int
	deleteCount  int
	startTime    time.Time
}

var DefaultSenderConfig = SenderConfig{
	QueueUrl:  "",
	BatchSize: pointer.Int(3),
}

// SenderConfig can be passed into NewSender() in order to configure
// the behavior of the sender.
type SenderConfig struct {
	QueueUrl  string `json:"queueUrl,omitempty"`
	BatchSize *int   `json:"batchSize,omitempty"`
}

type Sender struct {
	sync.Mutex
	sqsService *sqs.SQS
	config     SenderConfig
	count      int
	logger     zerolog.Logger
	eventBatch []event.Event
}

func (s *Sender) Unwrap() sender.Sender {
	return s
}
