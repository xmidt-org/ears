package syslog

import (
	"sync"

	"github.com/rs/zerolog"
	pkgplugin "github.com/xmidt-org/ears/pkg/plugin"
	"github.com/xmidt-org/ears/pkg/receiver"
	"github.com/xmidt-org/ears/pkg/tenant"
	"go.opentelemetry.io/otel/metric"
)

var _ receiver.Receiver = (*Receiver)(nil)

var (
	Name     = "syslog"
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
	)
}

type ReceiverConfig struct {
	Port string `json:"port"`
}

var DefaultReceiverConfig = ReceiverConfig{
	Port: "7531",
}

type Receiver struct {
	sync.Mutex
	logger              *zerolog.Logger
	config              ReceiverConfig
	name                string
	plugin              string
	tid                 tenant.Id
	eventSuccessCounter metric.BoundInt64Counter
	eventFailureCounter metric.BoundInt64Counter
	eventBytesCounter   metric.BoundInt64Counter
	next                receiver.NextFn
	syslogServer        *SyslogServer
}
