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

package s3

import (
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/rs/zerolog"
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
	Name     = "s3"
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
	Bucket:             "",
	Path:               "",
	Region:             endpoints.UsWest2RegionID,
	AcknowledgeTimeout: pointer.Int(5),
}

type ReceiverConfig struct {
	Bucket             string `json:"bucket,omitempty"`
	Path               string `json:"path,omitempty"`
	Region             string `json:"region,omitempty"`
	AcknowledgeTimeout *int   `json:"acknowledgeTimeout,omitempty"`
}

type Receiver struct {
	sync.Mutex
	s3Service           *s3.S3
	session             *session.Session
	done                chan struct{}
	stopped             bool
	config              ReceiverConfig
	name                string
	plugin              string
	tid                 tenant.Id
	next                receiver.NextFn
	logger              *zerolog.Logger
	count               int
	startTime           time.Time
	eventSuccessCounter metric.BoundInt64Counter
	eventFailureCounter metric.BoundInt64Counter
	eventBytesCounter   metric.BoundInt64Counter
}

var DefaultSenderConfig = SenderConfig{
	Bucket:   "",
	Path:     "",
	Region:   endpoints.UsWest2RegionID,
	FileName: "",
	FilePath: "",
}

// SenderConfig can be passed into NewSender() in order to configure
// the behavior of the sender.
type SenderConfig struct {
	Bucket   string `json:"bucket,omitempty"`
	Path     string `json:"path,omitempty"`
	Region   string `json:"region,omitempty"`
	FileName string `json:"fileName,omitempty"`
	FilePath string `json:"filePath,omitempty"`
}

type Sender struct {
	sync.Mutex
	s3Service           *s3.S3
	session             *session.Session
	name                string
	plugin              string
	tid                 tenant.Id
	config              SenderConfig
	count               int
	logger              *zerolog.Logger
	done                chan struct{}
	eventSuccessCounter metric.BoundInt64Counter
	eventFailureCounter metric.BoundInt64Counter
	eventBytesCounter   metric.BoundInt64Counter
	eventProcessingTime metric.BoundInt64Histogram
	eventSendOutTime    metric.BoundInt64Histogram
}
