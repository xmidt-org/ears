// SPDX-FileCopyrightText: 2025 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package nop

import (
	"github.com/rs/zerolog"
	pkgplugin "github.com/xmidt-org/ears/pkg/plugin"
	"github.com/xmidt-org/ears/pkg/tenant"
	"go.opentelemetry.io/otel/metric"

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
	pkgplugin.MetricPlugin
	done    chan struct{}
	stopped bool
	config  ReceiverConfig
	name    string
	plugin  string
	tid     tenant.Id
	next    receiver.NextFn
	logger  zerolog.Logger
}

var DefaultSenderConfig = SenderConfig{}

type SenderConfig struct {
}

type Sender struct {
	pkgplugin.MetricPlugin
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
