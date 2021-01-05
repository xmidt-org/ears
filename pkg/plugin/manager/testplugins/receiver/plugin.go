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

	"github.com/ghodss/yaml"
	"github.com/xmidt-org/ears/pkg/hasher"
	earsplugin "github.com/xmidt-org/ears/pkg/plugin"

	"github.com/xmidt-org/ears/pkg/receiver"
)

func main() {
	// required for `go build` to not fail
}

var Plugin = plugin{}

// for golangci-lint
var _ = Plugin

var _ earsplugin.NewPluginerer = (*plugin)(nil)
var _ earsplugin.Pluginer = (*plugin)(nil)
var _ receiver.NewReceiverer = (*plugin)(nil)
var _ receiver.Receiver = (*plugin)(nil)

// Plugin ============================================================

const (
	defaultPluginName    = "mock"
	defaultPluginVersion = "v0.0.0"
)

type PluginConfig struct {
	Name    string `yaml:"name"`
	Version string `yaml:"version"`

	source interface{}
}

type plugin struct {
	config PluginConfig
}

func (p *plugin) NewPluginer(config interface{}) (earsplugin.Pluginer, error) {
	return p.new(config)
}

func (p *plugin) PluginerHash(config interface{}) (string, error) {
	return hasher.Hash(config), nil
}

func (p *plugin) Name() string {
	return p.config.Name
}

func (p *plugin) Version() string {
	return p.config.Version
}

func (p *plugin) Config() string {
	cfg, err := yaml.Marshal(p.config)

	if err != nil {
		return "error: |\n  " + fmt.Sprint(err)
	}

	return string(cfg)
}

// Receiver ==========================================================

func (p *plugin) NewReceiver(config interface{}) (receiver.Receiver, error) {
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

// internal helpers ============================================================

func (p *plugin) new(config interface{}) (earsplugin.Pluginer, error) {
	cfg := PluginConfig{
		Name:    defaultPluginName,
		Version: defaultPluginVersion,
		source:  config,
	}

	c, ok := config.(string)
	if ok && c != "" {
		err := yaml.Unmarshal([]byte(c), &cfg)
		if err != nil {
			return nil, &earsplugin.InvalidConfigError{Err: err}
		}
	}

	return &plugin{config: cfg}, nil

}
