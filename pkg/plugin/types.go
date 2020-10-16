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

// === Plugin =========================================================

type Pluginer interface {
	Name() string
	Version() string
	Config() string
}

type NewPluginerer interface {
	NewPluginer(config string) (Pluginer, error)
}

// === Errors =========================================================

// ErrorCode holds a numeric value that can be used to communicate an
// enumerated list of error codes back to the caller.  It is expected
// that the user of the plugin will reference the plugin documentation
// to determine the meaning of the error.
type ErrorCode int

// Error wraps a plugin's errors.  Because this package will have
// no knowledge of which plugins it's importing, the convention is for the
// plugin to set a Code associated with the plugin.
type Error struct {
	Err    error
	Code   ErrorCode
	Values map[string]interface{}
}

type InvalidArgumentError struct {
	Err error
}

// NotSupportedError can be returned if a plugin doesn't support publishing
// subscribing.
type NotSupportedError struct{}
