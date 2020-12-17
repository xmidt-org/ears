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

package debug

import (
	"fmt"

	"github.com/xmidt-org/ears/pkg/hasher"
	pkgplugin "github.com/xmidt-org/ears/pkg/plugin"
	"github.com/xmidt-org/ears/pkg/receiver"
	"github.com/xmidt-org/ears/pkg/sender"
)

func (p *plugin) PluginerHash(config interface{}) (string, error) {
	return hasher.Hash(config), nil
}

func (p *plugin) NewPluginer(config interface{}) (pkgplugin.Pluginer, error) {
	plug := NewPlugin(p.name, p.version, p.commit)
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
	return fmt.Sprint(p.config)
}

func (p *plugin) ReceiverHash(config interface{}) (string, error) {
	return hasher.Hash(config), nil
}

func (p *plugin) NewReceiver(config interface{}) (receiver.Receiver, error) {
	return NewReceiver(config)
}

func (p *plugin) SenderHash(config interface{}) (string, error) {
	return hasher.Hash(config), nil
}

func (p *plugin) NewSender(config interface{}) (sender.Sender, error) {
	return NewSender(config)
}
