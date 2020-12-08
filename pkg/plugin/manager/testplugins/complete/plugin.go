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

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/pubsub/gochannel"
	"github.com/ghodss/yaml"
	"github.com/xmidt-org/ears/pkg/filter"
	"github.com/xmidt-org/ears/pkg/hasher"
	earsplugin "github.com/xmidt-org/ears/pkg/plugin"

	"github.com/xmidt-org/ears/pkg/event"
	"github.com/xmidt-org/ears/pkg/receiver"
	"github.com/xmidt-org/ears/pkg/sender"
)

var Plugin = plugin{}

var _ earsplugin.NewPluginerer = (*plugin)(nil)
var _ earsplugin.Pluginer = (*plugin)(nil)
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
	// ErrUnknown is returned when the error has not been properly categorized
	ErrUnknown earsplugin.ErrorCode = iota

	// ErrNotInitialized is when the plugin is not properly initialized
	ErrNotInitialized
)

// Plugin ============================================================

func (p *plugin) NewPluginer(config interface{}) (earsplugin.Pluginer, error) {
	return p.new(config)
}

func (p *plugin) PluginerHash(config interface{}) (string, error) {
	return hasher.Hash(config), nil
}

func (p *plugin) Name() string {
	return "managerTestPlugin"
}

func (p *plugin) Version() string {
	return "v0.0.0"
}

func (p *plugin) Config() string {
	c, err := yaml.Marshal(p.config)
	if err != nil {
		return fmt.Sprintf("config error: %w", err)
	}

	return string(c)
}

// Receiver ============================================================

func (p *plugin) NewReceiver(config interface{}) (receiver.Receiver, error) {
	if p.pubsub == nil {
		return nil, &earsplugin.Error{
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
		return nil, &earsplugin.Error{
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

// internal helpers ============================================================

func (p *plugin) new(config interface{}) (earsplugin.Pluginer, error) {
	p.config = config
	cfg := gochannel.Config{}

	c, ok := config.(string)
	if ok && c != "" {
		err := yaml.Unmarshal([]byte(c), &cfg)
		if err != nil {
			return nil, &earsplugin.InvalidConfigError{Err: err}
		}
	}

	return &plugin{
		pubsub: gochannel.NewGoChannel(
			cfg,
			watermill.NewStdLogger(false, false),
		),
	}, nil

}
