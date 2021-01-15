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

package match

import (
	"fmt"
	"strings"

	"github.com/goccy/go-yaml"
	"github.com/xmidt-org/ears/pkg/filter"
	"github.com/xmidt-org/ears/pkg/hasher"
	pkgplugin "github.com/xmidt-org/ears/pkg/plugin"
)

func (p *plugin) PluginerHash(config interface{}) (string, error) {
	return hasher.Hash(config), nil
}

func (p *plugin) NewPluginer(config interface{}) (pkgplugin.Pluginer, error) {
	plug := NewPlugin()
	plug.config = config

	return plug, nil
}

func (p *plugin) Name() string {
	return p.name
}

func (p *plugin) Version() string {
	return fmt.Sprintf("%s:%s", p.version, p.commit)
}

func (p *plugin) Config() string {
	if p.config == nil {
		return ""
	}

	out, err := yaml.Marshal(p.config)
	if err != nil {
		return fmt.Sprintf("could not export config: %s", err.Error())
	}

	return fmt.Sprint(strings.TrimSpace(string(out)))
}

func (p *plugin) FiltererHash(config interface{}) (string, error) {
	return hasher.Hash(config), nil
}

func (p *plugin) NewFilterer(config interface{}) (filter.Filterer, error) {
	return NewFilter(config)
}
