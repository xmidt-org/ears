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
	"github.com/xmidt-org/ears/pkg/plugin"
	"github.com/xmidt-org/ears/pkg/receiver"
	"github.com/xmidt-org/ears/pkg/sender"
)

type Config struct {
	Name   string `yaml:"name"`
	Path   string `yaml:"path"`
	Config string `yaml:"config"`
}

type Registration struct {
	Config Config
	Plugin plugin.Pluginer

	Capabilities Capabilities
}

type Capabilities struct {
	Receiver bool
	Filterer bool
	Sender   bool
}

type Manager interface {
	// Probably needs some sort of asset interface to
	// be able to load from file system, s3, and other places [Future]
	LoadPlugin(config Config) (plugin.Pluginer, error)

	RegisterPlugin(name string, p plugin.Pluginer) error
	UnregisterPlugin(name string) error

	Plugins() map[string]Registration
	Plugin(name string) Registration

	Receiverers() map[string]receiver.NewReceiverer
	NewReceiver(name string, config string) (receiver.Receiver, error)

	Filterers() map[string]filter.NewFilterer
	NewFilterer(name string, config string) (filter.Filterer, error)

	Senderers() map[string]sender.NewSenderer
	NewSender(name string, config string) (sender.Sender, error)
}

// === Errors =========================================================

// PluginLoadError is returned when the given plugin could
// not be loaded
type LoadError struct {
	Err error
}

// VariableLookupError is returned when no variable `Plugin`
// can be found in the compiled plugin
type VariableLookupError struct {
	Err error
}

// NewPluginerNotImplementedError is returned when the Plugin interface is
// not properly implemented
type NewPluginerNotImplementedError struct{}

type InvalidConfigError struct {
	Err error
}

type AlreadyRegisteredError struct{}

type NotFoundError struct{}
