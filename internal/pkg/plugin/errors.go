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

import "github.com/xmidt-org/ears/pkg/errs"

func (e *OptionError) Unwrap() error {
	return e.Err
}

func (e *OptionError) Error() string {
	return errs.String(
		"OptionError",
		map[string]interface{}{
			"message": e.Message,
		},
		e.Err,
	)
}

func (e *RegistrationError) Unwrap() error {
	return e.Err
}

func (e *RegistrationError) Error() string {
	return errs.String(
		"RegistrationError",
		map[string]interface{}{
			"message": e.Message,
			"plugin":  e.Name,
			"name":    e.Plugin,
		},
		e.Err,
	)
}

func (e *UnregistrationError) Unwrap() error {
	return e.Err
}

func (e *UnregistrationError) Error() string {
	return errs.String(
		"UnregistrationError",
		map[string]interface{}{
			"message": e.Message,
		},
		e.Err,
	)
}

func (e *NotRegisteredError) Unwrap() error {
	return nil
}

func (e *NotRegisteredError) Error() string {
	return errs.String("NotRegisteredError", nil, nil)
}
