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
	"fmt"

	"github.com/ghodss/yaml"
	earsplugin "github.com/xmidt-org/ears/pkg/plugin"
)

func main() {}

var Plugin = plugin{}

var _ earsplugin.NewPluginerer = (*plugin)(nil)
var _ earsplugin.Pluginer = (*plugin)(nil)

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

func (p *plugin) NewPluginer(config interface{}) (earsplugin.Pluginer, error) {
	return p.new(config)
}

func (p *plugin) PluginerHash(config interface{}) (string, error) {
	return "pluginerHash", nil
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

// internal helpers ============================================================

func (p *plugin) new(config interface{}) (earsplugin.Pluginer, error) {
	return nil, &earsplugin.InvalidConfigError{
		Err: fmt.Errorf("example config error"),
	}

}
