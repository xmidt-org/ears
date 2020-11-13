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

	pkgfilter "github.com/xmidt-org/ears/pkg/filter"
	pkgmanager "github.com/xmidt-org/ears/pkg/plugin/manager"
	pkgreceiver "github.com/xmidt-org/ears/pkg/receiver"
	pkgsender "github.com/xmidt-org/ears/pkg/sender"
)

type manager struct {
	pm             pkgmanager.Manager
	receivers      map[string]wrapper
	receiversCount map[string]int
	filters        map[string]wrapper
	filtersCount   map[string]int
	senders        map[string]wrapper
	sendersCount   map[string]int
}

func (m *manager) Receiverers() map[string]pkgreceiver.NewReceiverer {
	return m.pm.Receiverers()
}

func (m *manager) RegisterReceiver(pluginName string, config string) (pkgreceiver.Receiver, error) {
	return nil, &pkgmanager.NilPluginError{}
}
func (m *manager) UnregisterReceiver(pkgreceiver.Receiver) error {
	return &pkgmanager.NilPluginError{}
}

func (m *manager) Filterers() map[string]pkgfilter.NewFilterer {
	return m.pm.Filterers()
}

func (m *manager) RegisterFilter(pluginName string, config string) (pkgfilter.Filterer, error) {
	return nil, &pkgmanager.NilPluginError{}
}
func (m *manager) UnregisterFilter(pkgfilter.Filterer) error {
	return &pkgmanager.NilPluginError{}
}

func (m *manager) Senderers() map[string]pkgsender.NewSenderer {
	return m.pm.Senderers()
}

func (m *manager) RegisterSender(pluginName string, config string) (pkgsender.Sender, error) {

	ns, err := m.pm.Senderer(pluginName)
	if err != nil {
		return nil, err
	}

	hash, err := ns.SenderHash(config)
	if err != nil {
		return nil, fmt.Errorf("could not generate hash: %w", err)
	}

	w, ok := m.senders[hash]
	if ok {
		return w.sender, nil
	}

	s, err := ns.NewSender(config)
	if err != nil {
		return nil, fmt.Errorf("could not create sender: %w", err)
	}

	w = wrapper{
		manager:     m,
		wrapperType: typeWrapperSender,
		sender:      s,
		hash:        hash,
		active:      true,
	}

	m.senders[hash] = w
	m.sendersCount[hash]++

	return s, nil
}

func (m *manager) UnregisterSender(s pkgsender.Sender) error {
	return m.unregister(m.senders, m.sendersCount, s)
}

func (m *manager) unregister(mapping map[string]wrapper, count map[string]int, i interface{}) error {
	w, ok := i.(wrapper)
	if !ok || !w.active {
		return &pkgmanager.NotRegisteredError{}
	}

	w.active = false
	count[w.hash]--

	if count[w.hash] <= 0 {
		delete(mapping, w.hash)
		count[w.hash] = 0
	}

	return nil
}
