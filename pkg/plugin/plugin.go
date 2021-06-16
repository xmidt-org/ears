// Copyright 2021 Comcast Cable Communications Management, LLC
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
	"github.com/xmidt-org/ears/pkg/secret"
	"github.com/xmidt-org/ears/pkg/tenant"
	"strings"

	"github.com/xmidt-org/ears/pkg/bit"
	"github.com/xmidt-org/ears/pkg/filter"
	"github.com/xmidt-org/ears/pkg/hasher"
	"github.com/xmidt-org/ears/pkg/receiver"
	"github.com/xmidt-org/ears/pkg/sender"
	"gopkg.in/yaml.v2"
)

// == Options ===========================================================

func (p *Plugin) WithName(name string) error {
	p.name = name
	return nil

}
func (p *Plugin) WithVersion(version string) error    { p.version = version; return nil }
func (p *Plugin) WithCommitID(commitID string) error  { p.commitID = commitID; return nil }
func (p *Plugin) WithConfig(config interface{}) error { p.config = config; return nil }

func (p *Plugin) WithNewPluginer(fn NewPluginerFn) error {
	if p == nil {
		return &NilPluginError{}
	}

	if fn == nil {
		return fmt.Errorf("nil NewPluginerFn provided")
	}

	p.Lock()
	defer p.Unlock()
	p.newPluginer = fn
	p.types.Set(TypePluginer)

	return nil
}

func (p *Plugin) WithPluginerHasher(fn HashFn) error {
	if p == nil {
		return &NilPluginError{}
	}

	return p.setHasher(&p.hashPluginer, fn)
}

func (p *Plugin) WithNewFilterer(fn NewFiltererFn) error {
	if p == nil {
		return &NilPluginError{}
	}

	if fn == nil {
		return fmt.Errorf("nil NewFiltererFn provided")
	}

	p.Lock()
	defer p.Unlock()
	p.newFilterer = fn
	p.types.Set(TypeFilter)

	return nil
}

func (p *Plugin) WithFilterHasher(fn HashFn) error {
	if p == nil {
		return &NilPluginError{}
	}

	return p.setHasher(&p.hashFilter, fn)
}

func (p *Plugin) WithNewReceiver(fn NewReceiverFn) error {
	if p == nil {
		return &NilPluginError{}
	}

	if fn == nil {
		return fmt.Errorf("nil NewReceiverFn provided")
	}

	p.Lock()
	defer p.Unlock()
	p.newReceiver = fn
	p.types.Set(TypeReceiver)

	return nil
}
func (p *Plugin) WithReceiverHasher(fn HashFn) error {
	return p.setHasher(&p.hashReceiver, fn)
}

func (p *Plugin) WithNewSender(fn NewSenderFn) error {

	if p == nil {
		return &NilPluginError{}
	}

	if fn == nil {
		return fmt.Errorf("nil NewSenderFn provided")
	}

	p.Lock()
	defer p.Unlock()
	p.newSender = fn
	p.types.Set(TypeSender)

	return nil
}
func (p *Plugin) WithSenderHasher(fn HashFn) error {
	if p == nil {
		return &NilPluginError{}
	}

	return p.setHasher(&p.hashSender, fn)
}

func (p *Plugin) SupportedTypes() bit.Mask {
	if p == nil {
		return bit.Mask(0)
	}

	return p.types
}

// == Construction =======================================================

func NewPlugin(options ...Option) (*Plugin, error) {

	defaultHasher := func(i interface{}) (string, error) {
		return hasher.Hash(i), nil
	}

	p := Plugin{
		hashPluginer: defaultHasher,
		hashReceiver: defaultHasher,
		hashFilter:   defaultHasher,
		hashSender:   defaultHasher,
	}

	var err error
	for _, option := range options {
		err = option(&p)
		if err != nil {
			return nil, &OptionError{Err: err}
		}
	}

	return &p, nil

}

// == Pluginer ===========================================================

func (p *Plugin) PluginerHash(config interface{}) (string, error) {
	if p == nil || p.hashPluginer == nil {
		return "", &NilPluginError{}
	}

	return p.hashPluginer(config)
}

func (p *Plugin) NewPluginer(config interface{}) (Pluginer, error) {
	if p == nil {
		return nil, &NilPluginError{}
	}

	if p.newPluginer == nil {
		return p, nil
	}

	return p.newPluginer(config)
}

func (p *Plugin) Name() string {
	return p.name
}

func (p *Plugin) Version() string {
	return p.version
}

func (p *Plugin) CommitID() string {
	return p.commitID
}

func (p *Plugin) Config() string {
	if p.config == nil {
		return ""
	}

	out, err := yaml.Marshal(p.config)
	if err != nil {
		return fmt.Sprintf("could not export config: %s", err.Error())
	}

	return fmt.Sprint(strings.TrimSpace(string(out)))
}

// == Receiverer ===========================================================

func (p *Plugin) ReceiverHash(config interface{}) (string, error) {
	if p == nil || p.hashReceiver == nil {
		return "", &NilPluginError{}
	}

	return p.hashReceiver(config)
}

func (p *Plugin) NewReceiver(tid tenant.Id, plugin string, name string, config interface{}, secrets secret.Vault) (receiver.Receiver, error) {
	if p == nil {
		return nil, &NilPluginError{}
	}

	if p.newReceiver == nil {
		return nil, &NotSupportedError{}
	}

	return p.newReceiver(tid, plugin, name, config, secrets)
}

// == Senderer ===========================================================

func (p *Plugin) SenderHash(config interface{}) (string, error) {
	if p == nil || p.hashSender == nil {
		return "", &NilPluginError{}
	}

	return p.hashSender(config)
}

func (p *Plugin) NewSender(tid tenant.Id, plugin string, name string, config interface{}, secrets secret.Vault) (sender.Sender, error) {
	if p == nil {
		return nil, &NilPluginError{}
	}

	if p.newSender == nil {
		return nil, &NotSupportedError{}
	}

	return p.newSender(tid, plugin, name, config, secrets)
}

// == Filterer ===========================================================

func (p *Plugin) FiltererHash(config interface{}) (string, error) {
	if p == nil || p.hashFilter == nil {
		return "", &NilPluginError{}
	}

	return p.hashFilter(config)
}

func (p *Plugin) NewFilterer(tid tenant.Id, plugin string, name string, config interface{}, secrets secret.Vault) (filter.Filterer, error) {
	if p == nil {
		return nil, &NilPluginError{}
	}

	if p.newFilterer == nil {
		return nil, &NotSupportedError{}
	}

	return p.newFilterer(tid, plugin, name, config, secrets)
}

// == Helpers =================================

func (p *Plugin) setHasher(field *HashFn, fn HashFn) error {

	if fn == nil {
		return fmt.Errorf("nil Hash Function provided")
	}

	if p == nil {
		return &NilPluginError{}
	}

	if p == nil {
		return &NilPluginError{}
	}

	p.Lock()
	defer p.Unlock()
	*field = fn

	return nil

}
