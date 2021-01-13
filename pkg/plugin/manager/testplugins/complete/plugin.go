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
	"fmt"

	"github.com/ThreeDotsLabs/watermill/pubsub/gochannel"
	"github.com/xmidt-org/ears/pkg/filter"
	"github.com/xmidt-org/ears/pkg/hasher"
	pkgplugin "github.com/xmidt-org/ears/pkg/plugin"

	"github.com/xmidt-org/ears/pkg/event"
	"github.com/xmidt-org/ears/pkg/receiver"
	"github.com/xmidt-org/ears/pkg/sender"
)

func main() {
	// required for `go build` to not fail
}

var Plugin = plugin{}

// for golangci-lint
var _ = Plugin

var _ pkgplugin.NewPluginerer = (*plugin)(nil)
var _ sender.NewSenderer = (*plugin)(nil)
var _ sender.Sender = (*plugin)(nil)
var _ receiver.NewReceiverer = (*plugin)(nil)
var _ receiver.Receiver = (*plugin)(nil)
var _ filter.NewFilterer = (*plugin)(nil)
var _ filter.Filterer = (*plugin)(nil)

type plugin struct {
	pubsub *gochannel.GoChannel
	config interface{}
}

// == Custom Error Codes ==========================================================

const (
	// // ErrUnknown is returned when the error has not been properly categorized
	// ErrUnknown earsplugin.ErrorCode = iota

	// ErrNotInitialized is when the plugin is not properly initialized
	ErrNotInitialized pkgplugin.ErrorCode = iota + 1
)

// Pluginer ============================================================

func (p *plugin) NewPluginer(config interface{}) (pkgplugin.Pluginer, error) {
	return pkgplugin.NewPlugin(
		pkgplugin.WithName("name"),
		pkgplugin.WithVersion("version"),
		pkgplugin.WithCommitID("commitId"),
		pkgplugin.WithConfig(config),
	)
}

func (p *plugin) PluginerHash(config interface{}) (string, error) {
	return "pluginerHash", nil
}

// Receiver ============================================================

func (p *plugin) NewReceiver(config interface{}) (receiver.Receiver, error) {
	if p.pubsub == nil {
		return nil, &pkgplugin.Error{
			Code: ErrNotInitialized,
			Err:  fmt.Errorf("NotInitialized"),
		}
	}

	return p, nil
}

func (p *plugin) ReceiverHash(config interface{}) (string, error) {
	return hasher.Hash(config), nil
}

func (p *plugin) Receive(ctx context.Context, next receiver.NextFn) error {
	return nil
}

func (p *plugin) StopReceiving(ctx context.Context) error {
	return nil
}

// Filterer ============================================================

func (p *plugin) NewFilterer(config interface{}) (filter.Filterer, error) {
	return p, nil
}

func (p *plugin) FiltererHash(config interface{}) (string, error) {
	return hasher.Hash(config), nil
}

func (p *plugin) Filter(ctx context.Context, e event.Event) ([]event.Event, error) {
	return nil, nil
}

// Sender ============================================================

func (p *plugin) NewSender(config interface{}) (sender.Sender, error) {
	if p.pubsub == nil {
		return nil, &pkgplugin.Error{
			Code: ErrNotInitialized,
			Err:  fmt.Errorf("NotInitialized"),
		}
	}

	return p, nil
}

func (p *plugin) SenderHash(config interface{}) (string, error) {
	return hasher.Hash(config), nil
}

func (p *plugin) Send(ctx context.Context, event event.Event) error {
	return nil
}
