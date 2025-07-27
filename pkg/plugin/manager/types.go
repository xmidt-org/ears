// SPDX-FileCopyrightText: 2025 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package manager

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
	Receiverer(pluginName string) (receiver.NewReceiverer, error)

	Filterers() map[string]filter.NewFilterer
	Filterer(pluginName string) (filter.NewFilterer, error)

	Senderers() map[string]sender.NewSenderer
	Senderer(pluginName string) (sender.NewSenderer, error)
}

// === Errors =========================================================

// OpenPluginError is returned when the Go plugin framework
// is unable to load the plugin.
type OpenPluginError struct {
	Err error
}

// NewPluginerError is returned when the call to NewPluginer
// on a loaded plugin results in an error.
type NewPluginerError struct {
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

type NewSendererNotImplementedError struct{}
type NewReceivererNotImplementedError struct{}
type NewFiltererNotImplementedError struct{}

type InvalidConfigError struct {
	Err error
}

type AlreadyRegisteredError struct{}

type NotFoundError struct{}

type NilPluginError struct{}

type NotRegisteredError struct{}
