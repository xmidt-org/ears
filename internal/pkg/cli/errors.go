/**
 *  Copyright (c) 2020  Comcast Cable Communications Management, LLC
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package cli

import (
	"fmt"
	"strings"
)

const (
	ErrConfigNotSupportedProtocol = "ConfigNotSupportedProtocol"
)

// argError =============================================================

type ArgError struct {
	Err   error
	key   string
	value interface{}
}

func (a ArgError) Key() string {
	return a.key
}

func (a ArgError) Value() interface{} {
	return a.value
}

func (a ArgError) Error() string {
	msg := "key=" + a.key
	if a.value != nil {
		msg += fmt.Sprintf(", value=%v", a.value)
	}

	if a.Err != nil {
		msg = a.Err.Error() + ", " + msg
	}
	return msg
}

func (a *ArgError) Unwrap() error {
	return a.Err
}

// configError =============================================================

type ConfigError struct {
	Err  error
	path string
}

func (c ConfigError) Path() string {
	return c.path
}

func (c ConfigError) Error() string {
	msg := ""
	if c.Err != nil {
		// In case there are newlines in the error text
		msg = strings.Replace(c.Err.Error(), "\n", " ", -1)
	}

	if msg != "" {
		msg += ", "
	}

	return msg + "path: " + c.Path()
}

func (c *ConfigError) Unwrap() error {
	return c.Err
}
