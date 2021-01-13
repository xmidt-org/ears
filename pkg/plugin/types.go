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
	"sync"

	"github.com/xmidt-org/ears/pkg/filter"
	"github.com/xmidt-org/ears/pkg/receiver"
	"github.com/xmidt-org/ears/pkg/sender"
)

// === Plugin =========================================================

//go:generate rm -f testing_mock.go
//go:generate moq -out testing_mock.go . Hasher NewPluginerer Pluginer

// Hasher defines the hashing interface that a pluginer
// needs to implement
type Hasher interface {
	// PluginerHash calculates the hash of a receiver based on the
	// given configuration
	PluginerHash(config interface{}) (string, error)
}

// NewPluginerer is something that is able to return a new Pluginer
type NewPluginerer interface {
	Hasher
	NewPluginer(config interface{}) (Pluginer, error)
}

// Pluginer defines values that should be returned in order to
// get visibility into which plugin was loaded and how it was
// configured.
type Pluginer interface {
	Name() string
	Version() string
	CommitID() string
	Config() string

	// A bitmask of supported types
	SupportedTypes() Bitmask
}

type Bitmask uint

const (
	TypePluginer Bitmask = 1 << iota
	TypeReceiver
	TypeSender
	TypeFilter
)

type HashFn func(i interface{}) (string, error)

// TODO: Question -- How do these stay in sync with the package?
// Do we define them in the package alongside the interface definition?
type NewPluginerFn func(config interface{}) (Pluginer, error)
type NewReceiverFn func(config interface{}) (receiver.Receiver, error)
type NewSenderFn func(config interface{}) (sender.Sender, error)
type NewFiltererFn func(config interface{}) (filter.Filterer, error)

// Plugin implements Pluginer
type Plugin struct {
	sync.Mutex

	name     string
	version  string
	commitID string
	config   interface{}

	types Bitmask

	hashPluginer HashFn
	hashReceiver HashFn
	hashFilter   HashFn
	hashSender   HashFn

	newPluginer NewPluginerFn
	newReceiver NewReceiverFn
	newSender   NewSenderFn
	newFilterer NewFiltererFn
}

// === Errors =========================================================

// ErrorCode holds a numeric value that can be used to communicate an
// enumerated list of error codes back to the caller.  It is expected
// that the user of the plugin will reference the plugin documentation
// to determine the meaning of the error.
type ErrorCode int

// TODO:  Define an enumeration of codes to use to help determine
// how to handle the error

// Error wraps a plugin's errors.  Because this package will have
// no knowledge of which plugins it's importing, the convention is for the
// plugin to set a Code associated with the plugin.
type Error struct {
	Err  error
	Code ErrorCode
}

// InvalidConfigError is returned when a configuration parameter
// results in a plugin error
type InvalidConfigError struct {
	Err error
}

// NotSupportedError can be returned if a plugin doesn't support publishing
// subscribing.
type NotSupportedError struct{}

type NilPluginError struct{}
