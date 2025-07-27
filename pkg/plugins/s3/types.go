// SPDX-FileCopyrightText: 2025 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package s3

import (
	"time"

	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/rs/zerolog"
	"github.com/xmidt-org/ears/pkg/errs"
	"github.com/xmidt-org/ears/pkg/secret"
	"github.com/xmidt-org/ears/pkg/tenant"
	"github.com/xorcare/pointer"
	"go.opentelemetry.io/otel/metric"

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
	AWSRegion:          endpoints.UsWest2RegionID,
	AWSRoleARN:         "",
	AWSSecretAccessKey: "",
	AWSAccessKeyId:     "",
	AcknowledgeTimeout: pointer.Int(5),
}

type ReceiverConfig struct {
	Bucket             string `json:"bucket,omitempty"`
	Path               string `json:"path,omitempty"`
	AcknowledgeTimeout *int   `json:"acknowledgeTimeout,omitempty"`
	AWSRoleARN         string `json:"awsRoleARN,omitempty"`
	AWSAccessKeyId     string `json:"awsAccessKeyId,omitempty"`
	AWSSecretAccessKey string `json:"awsSecretAccessKey,omitempty"`
	AWSRegion          string `json:"awsRegion,omitempty"`
}

type Receiver struct {
	pkgplugin.MetricPlugin
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
	secrets             secret.Vault
	startTime           time.Time
	eventSuccessCounter metric.BoundInt64Counter
	eventFailureCounter metric.BoundInt64Counter
	eventBytesCounter   metric.BoundInt64Counter
	awsRoleArn          string
	awsAccessKey        string
	awsAccessSecret     string
	awsRegion           string
	bucket              string
}

var DefaultSenderConfig = SenderConfig{
	Bucket:             "",
	Path:               "",
	AWSRegion:          endpoints.UsWest2RegionID,
	AWSRoleARN:         "",
	AWSSecretAccessKey: "",
	AWSAccessKeyId:     "",
	FileName:           "",
	FilePath:           "",
}

// SenderConfig can be passed into NewSender() in order to configure
// the behavior of the sender.
type SenderConfig struct {
	Bucket             string `json:"bucket,omitempty"`
	Path               string `json:"path,omitempty"`
	FileName           string `json:"fileName,omitempty"`
	FilePath           string `json:"filePath,omitempty"`
	AWSRoleARN         string `json:"awsRoleARN,omitempty"`
	AWSAccessKeyId     string `json:"awsAccessKeyId,omitempty"`
	AWSSecretAccessKey string `json:"awsSecretAccessKey,omitempty"`
	AWSRegion          string `json:"awsRegion,omitempty"`
}

type Sender struct {
	pkgplugin.MetricPlugin
	s3Service           *s3.S3
	session             *session.Session
	name                string
	plugin              string
	tid                 tenant.Id
	config              SenderConfig
	logger              *zerolog.Logger
	done                chan struct{}
	secrets             secret.Vault
	eventSuccessCounter metric.BoundInt64Counter
	eventFailureCounter metric.BoundInt64Counter
	eventBytesCounter   metric.BoundInt64Counter
	eventProcessingTime metric.BoundInt64Histogram
	eventSendOutTime    metric.BoundInt64Histogram
	awsRoleArn          string
	awsAccessKey        string
	awsAccessSecret     string
	awsRegion           string
	bucket              string
}

type S3Error struct {
	op  string
	err error
}

func (e *S3Error) Error() string {
	return errs.String("S3Error", map[string]interface{}{"op": e.op}, e.err)
}

func (e *S3Error) Unwrap() error {
	return e.err
}
