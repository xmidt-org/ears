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

package route

type (
	MissingFilterPluginConfigError struct {
	}
	UnknownFilterTypeError struct {
		FilterType string
	}
	EmptyPluginHashError struct {
	}
	EmptyHashError struct {
	}
	MissingRouteError struct {
	}
	UnkownRouteError struct {
	}
	UnknownPluginTypeError struct {
		PluginType string
	}
	MissingPluginConfiguratonError struct {
		PluginType string
		PluginHash string
		PluginMode string
	}
	UnworthyPluginError struct {
	}
)

func (e *MissingPluginConfiguratonError) Error() string {
	return "missing configuration for " + e.PluginType + " " + e.PluginMode + " plugin " + e.PluginHash
}

func (e *UnknownPluginTypeError) Error() string {
	return "unknown plugin type " + e.PluginType
}

func (e *MissingFilterPluginConfigError) Error() string {
	return "missing filter plugin config"
}

func (e *UnknownFilterTypeError) Error() string {
	return "unknown filter type " + e.FilterType
}

func (e *EmptyPluginHashError) Error() string {
	return "empty plugin hash"
}

func (e *EmptyHashError) Error() string {
	return "empty hash"
}

func (e *MissingRouteError) Error() string {
	return "missing routing table entry"
}

func (e *UnkownRouteError) Error() string {
	return "unknown route"
}

func (e *UnworthyPluginError) Error() string {
	return "unworthy plugin"
}
