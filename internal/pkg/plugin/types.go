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
	"github.com/xmidt-org/ears/pkg/filter"
	"github.com/xmidt-org/ears/pkg/receiver"
	"github.com/xmidt-org/ears/pkg/sender"
)

type Manager interface {
	Receiverers() map[string]receiver.NewReceiverer
	RegisterReceiver(pluginName string, config string) (receiver.Receiver, error)
	UnregisterReceiver(receiver.Receiver) error

	Filterers() map[string]filter.NewFilterer
	RegisterFilter(pluginName string, config string) (filter.Filterer, error)
	UnregisterFilter(filter.Filterer) error

	Senderers() map[string]sender.NewSenderer
	RegisterSender(pluginName string, config string) (sender.Sender, error)
	UnregisterSender(filter.Filterer) error
}
