// Licensed to Comcast Cable Communications Management, LLC under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. Comcast Cable Communications Management, LLC licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package main

import (
	"context"
	"fmt"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/pubsub/gochannel"
	"github.com/ghodss/yaml"
	earsplugin "github.com/xmidt-org/ears/pkg/plugin"
)

var Plugin = plugin{}

var _ earsplugin.Sender = (*plugin)(nil)
var _ earsplugin.Receiver = (*plugin)(nil)

type plugin struct {
	pubsub *gochannel.GoChannel
	config string
}

// == Custom Error Codes ==========================================================

const (
	// ErrUnknown is returned when the error has not been properly categorized
	ErrUnknown earsplugin.ErrorCode = iota

	// ErrNotInitialized is when the plugin is not properly initialized
	ErrNotInitialized
)

// Plugin ============================================================

func (p *plugin) Name() string {
	return "managerTestPlugin"
}

func (p *plugin) Version() string {
	return "v0.0.0"
}

func (p *plugin) Config() string {
	return p.config
}

// Receiver ============================================================

func (p *plugin) NewReceiver(config string) (earsplugin.Receiver, error) {
	if p.pubsub == nil {
		return nil, &earsplugin.Error{
			Code: ErrNotInitialized,
			Err:  fmt.Errorf("NotInitialized"),
		}
	}

	return p, nil
}

func (p *plugin) Receive(ctx context.Context, next earsplugin.NextFn) error {
	return nil
}

// Sender ============================================================

func (p *plugin) NewSender(config string) (earsplugin.Sender, error) {
	if p.pubsub == nil {
		return nil, &earsplugin.Error{
			Code: ErrNotInitialized,
			Err:  fmt.Errorf("NotInitialized"),
		}
	}

	return p, nil
}

func (p *plugin) Send(ctx context.Context, event earsplugin.Event) error {
	return nil
}

// internal helpers ============================================================

func (p *plugin) new(config string) (earsplugin.Plugin, error) {
	p.config = config
	cfg := gochannel.Config{}

	if config != "" {
		err := yaml.Unmarshal([]byte(config), &cfg)
		if err != nil {
			return nil, &earsplugin.InvalidArgumentError{Err: err}
		}
	}

	return &plugin{
		pubsub: gochannel.NewGoChannel(
			cfg,
			watermill.NewStdLogger(false, false),
		),
	}, nil

}
