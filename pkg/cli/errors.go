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
	"github.com/xmidt-org/ears/pkg/errs"
)

type EmptyCmdArgumentError struct {
}

func (e *EmptyCmdArgumentError) Error() string {
	return errs.String("EmptyCmdArgumentError", nil, nil)
}

type NoDefaultSetError struct {
}

func (e *NoDefaultSetError) Error() string {
	return errs.String("NoDefaultSetError", nil, nil)
}

type ConfigNotSupportedProtocolError struct {
}

func (e *ConfigNotSupportedProtocolError) Error() string {
	return errs.String("ConfigNotSupportedProtocolError", nil, nil)
}

type ConfigError struct {
	Err  error
	path string
}

func (c *ConfigError) Error() string {
	return errs.String("ConfigError", map[string]interface{}{"path": c.path}, c.Err)
}

func (c *ConfigError) Unwrap() error {
	return c.Err
}
