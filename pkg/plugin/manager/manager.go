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

package plugin

import (
	"fmt"
	"sync"

	goplugin "plugin"

	"github.com/xmidt-org/ears/pkg/filter"
	"github.com/xmidt-org/ears/pkg/plugin"
	"github.com/xmidt-org/ears/pkg/receiver"
	"github.com/xmidt-org/ears/pkg/sender"
)

var _ Manager = (*manager)(nil)

type manager struct {
	sync.Mutex
	pathPrefix string

	// Map of the name used in the configuration to the loaded plugin
	registrations map[string]Registration
}

type Option func(OptionProcessor) error
type OptionProcessor interface{}

func NewManager(options ...Option) *manager {
	return &manager{
		registrations: map[string]Registration{},
	}
}

// === Registration ===================================================

func (m *manager) LoadPlugin(config Config) (plugin.Pluginer, error) {

	if config.Name == "" {
		return nil, &InvalidConfigError{
			Err: fmt.Errorf("config name cannot be empty"),
		}
	}

	if config.Path == "" {
		return nil, &InvalidConfigError{
			Err: fmt.Errorf("config path cannot be empty"),
		}
	}

	library, err := goplugin.Open(config.Path)
	if err != nil {
		return nil, &LoadError{
			Err: fmt.Errorf("could not open plugin: %w", err),
		}
	}

	newerVar, err := library.Lookup("Plugin")
	if err != nil {
		return nil, &VariableLookupError{
			Err: fmt.Errorf("could not find Plugin variable: %w", err),
		}
	}

	newer, ok := newerVar.(plugin.NewPluginerer)
	if !ok {
		return nil, &NewPluginerNotImplementedError{}
	}

	plug, err := newer.NewPluginer(config.Config)
	if err != nil {
		return nil, &LoadError{
			Err: fmt.Errorf("could not create plugin: %w", err),
		}
	}

	// m.registrations[config.Name] = Registration{
	// 	Config: config,
	// 	Plugin: plug,
	// }

	// TODO:  How do we not lose the other interesting info in
	// regard to config path and initialization data?
	// Maybe copy the _ img import process & etc.
	return plug, m.RegisterPlugin(config.Name, plug)
}

func (m *manager) RegisterPlugin(name string, p plugin.Pluginer) error {
	_, isReceiver := p.(receiver.NewReceiverer)
	_, isFilterer := p.(filter.NewFilterer)
	_, isSender := p.(sender.NewSenderer)

	m.Lock()
	defer m.Unlock()

	if _, ok := m.registrations[name]; ok {
		return &AlreadyRegisteredError{}
	}

	m.registrations[name] = Registration{
		Config: Config{Name: name},
		Plugin: p,
		Capabilities: Capabilities{
			Receiver: isReceiver,
			Filterer: isFilterer,
			Sender:   isSender,
		},
	}

	return nil
}

func (m *manager) UnregisterPlugin(name string) error {
	m.Lock()
	defer m.Unlock()

	delete(m.registrations, name)
	return nil
}

// === Plugins ========================================================

func (m *manager) Plugins() map[string]Registration {
	m.Lock()
	defer m.Unlock()

	r := make(map[string]Registration, len(m.registrations))
	for k, v := range m.registrations {
		r[k] = v
	}

	return r
}

func (m *manager) Plugin(name string) Registration {
	r, _ := m.registrations[name]
	return r
}

// === Senders ========================================================

func (m *manager) Senderers() map[string]sender.NewSenderer {
	hash := map[string]sender.NewSenderer{}
	m.Lock()
	defer m.Unlock()

	for k, r := range m.registrations {
		p, ok := r.Plugin.(sender.NewSenderer)
		if ok {
			hash[k] = p
		}
	}

	return hash
}

func (m *manager) NewSender(name string, config string) (sender.Sender, error) {

	p, ok := m.registrations[name].Plugin.(sender.NewSenderer)

	if !ok {
		return nil, &NotFoundError{}
	}

	return p.NewSender(config)

}

// === Filters ========================================================

func (m *manager) Filterers() map[string]filter.NewFilterer {
	hash := map[string]filter.NewFilterer{}
	m.Lock()
	defer m.Unlock()

	for k, r := range m.registrations {
		p, ok := r.Plugin.(filter.NewFilterer)
		if ok {
			hash[k] = p
		}
	}

	return hash
}

func (m *manager) NewFilterer(name string, config string) (filter.Filterer, error) {

	p, ok := m.registrations[name].Plugin.(filter.NewFilterer)

	if !ok {
		return nil, &NotFoundError{}
	}

	return p.NewFilterer(config)

}

// === Receivers ======================================================

func (m *manager) Receiverers() map[string]receiver.NewReceiverer {
	hash := map[string]receiver.NewReceiverer{}
	m.Lock()
	defer m.Unlock()

	for k, r := range m.registrations {
		p, ok := r.Plugin.(receiver.NewReceiverer)
		if ok {
			hash[k] = p
		}
	}

	return hash
}

func (m *manager) NewReceiver(name string, config string) (receiver.Receiver, error) {

	p, ok := m.registrations[name].Plugin.(receiver.NewReceiverer)

	if !ok {
		return nil, &NotFoundError{}
	}

	return p.NewReceiver(config)

}
