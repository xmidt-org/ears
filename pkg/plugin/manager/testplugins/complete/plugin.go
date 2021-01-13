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

package main

import (
	"context"

	"github.com/xmidt-org/ears/pkg/filter"
	pkgplugin "github.com/xmidt-org/ears/pkg/plugin"

	"github.com/xmidt-org/ears/pkg/event"
	"github.com/xmidt-org/ears/pkg/receiver"
	"github.com/xmidt-org/ears/pkg/sender"
)

func main() {
	// required for `go build` to not fail
}

var Plugin, PluginErr = NewPlugin()

// for golangci-lint
var _ = Plugin
var _ = PluginErr

var _ sender.Sender = (*plugin)(nil)
var _ receiver.Receiver = (*plugin)(nil)
var _ filter.Filterer = (*plugin)(nil)

type plugin struct{}

// == Custom Error Codes ==========================================================

const (
	// // ErrUnknown is returned when the error has not been properly categorized
	// ErrUnknown earsplugin.ErrorCode = iota

	// ErrNotInitialized is when the plugin is not properly initialized
	ErrNotInitialized pkgplugin.ErrorCode = iota + 1
)

const (
	Name     = "name"
	Version  = "version"
	CommitID = "commitID"
)

// ===================================================

func NewPlugin() (*pkgplugin.Plugin, error) {
	return NewPluginVersion(Name, Version, CommitID)
}

func NewPluginVersion(name string, version string, commitID string) (*pkgplugin.Plugin, error) {
	return pkgplugin.NewPlugin(
		pkgplugin.WithName(name),
		pkgplugin.WithVersion(version),
		pkgplugin.WithCommitID(commitID),
		pkgplugin.WithNewFilterer(NewFilterer),
		pkgplugin.WithNewReceiver(NewReceiver),
		pkgplugin.WithNewSender(NewSender),
	)
}

// Receiver ============================================================

func NewReceiver(config interface{}) (receiver.Receiver, error) {
	return &plugin{}, nil
}

func (p *plugin) Receive(ctx context.Context, next receiver.NextFn) error {
	return nil
}

func (p *plugin) StopReceiving(ctx context.Context) error {
	return nil
}

// Filterer ============================================================

func NewFilterer(config interface{}) (filter.Filterer, error) {
	return &plugin{}, nil
}

func (p *plugin) Filter(ctx context.Context, e event.Event) ([]event.Event, error) {
	return nil, nil
}

// Sender ============================================================

func NewSender(config interface{}) (sender.Sender, error) {
	return &plugin{}, nil
}

func (p *plugin) Send(ctx context.Context, event event.Event) error {
	return nil
}
