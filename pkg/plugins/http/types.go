package http

import (
	"github.com/rs/zerolog"
	"github.com/xmidt-org/ears/pkg/errs"
	pkgplugin "github.com/xmidt-org/ears/pkg/plugin"
	"github.com/xmidt-org/ears/pkg/receiver"
	"github.com/xmidt-org/ears/pkg/sender"
	"net/http"
)

var _ sender.Sender = (*Sender)(nil)
var _ receiver.Receiver = (*Receiver)(nil)

var (
	Name     = "http"
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

type ReceiverConfig struct {
	Path   string `json:"path,omitempty"`
	Port   *int   `json:"port,omitempty"`
	Method string `json:"method:omitempty"`
}

type Receiver struct {
	path   string
	method string
	port   int
	logger *zerolog.Logger
	srv    *http.Server
}

type SenderConfig struct {
	Url    string `json:"url,omitempty"`
	Method string `json:"method,omitempty"`
}

type Sender struct {
	client *http.Client
	method string
	url    string
}

type BadHttpStatusError struct {
	statusCode int
}

func (e *BadHttpStatusError) Error() string {
	return errs.String("BadHttpStatusError", map[string]interface{}{"statusCode": e.statusCode}, nil)
}
