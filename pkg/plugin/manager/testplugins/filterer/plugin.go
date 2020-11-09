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
	"crypto/md5"
	"encoding/hex"
	"fmt"

	"github.com/ghodss/yaml"
	"github.com/xmidt-org/ears/pkg/filter"
	"github.com/xmidt-org/ears/pkg/hasher"
	earsplugin "github.com/xmidt-org/ears/pkg/plugin"

	"github.com/xmidt-org/ears/pkg/event"
)

var Plugin = plugin{}

var _ earsplugin.NewPluginerer = (*plugin)(nil)
var _ earsplugin.Pluginer = (*plugin)(nil)
var _ filter.NewFilterer = (*plugin)(nil)
var _ filter.Filterer = (*plugin)(nil)

// == Custom Error Codes =============================================

const (
	// ErrUnknown is returned when the error has not been properly
	// categorized
	ErrUnknown earsplugin.ErrorCode = iota

	// ErrNotInitialized is when the plugin is not properly initialized
	ErrNotInitialized
)

// Plugin ============================================================

const (
	defaultPluginName    = "mock"
	defaultPluginVersion = "v0.0.0"
)

type PluginConfig struct {
	Name    string `yaml:"name"`
	Version string `yaml:"version"`

	source string
}

type plugin struct {
	config PluginConfig
}

func (p *plugin) NewPluginer(config string) (earsplugin.Pluginer, error) {
	return p.new(config)
}

func (p *plugin) PluginerHash(config string) (string, error) {
	hash := md5.Sum([]byte(config))
	return hex.EncodeToString(hash[:]), nil
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

// Filterer ============================================================

func (p *plugin) NewFilterer(config string) (filter.Filterer, error) {
	return p, nil
}

func (p *plugin) FiltererHash(config string) (string, error) {
	return hasher.Hash(config), nil
}

func (p *plugin) Filter(ctx context.Context, e event.Event) ([]event.Event, error) {
	return nil, nil
}

// internal helpers ============================================================

func (p *plugin) new(config string) (earsplugin.Pluginer, error) {
	cfg := PluginConfig{
		Name:    defaultPluginName,
		Version: defaultPluginVersion,
		source:  config,
	}

	if config != "" {
		err := yaml.Unmarshal([]byte(config), &cfg)
		if err != nil {
			return nil, &earsplugin.InvalidConfigError{Err: err}
		}
	}

	return &plugin{config: cfg}, nil

}
