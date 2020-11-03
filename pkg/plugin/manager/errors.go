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

package manager

import "fmt"

func (e *OpenPluginError) Unwrap() error {
	return e.Err
}

func (e *OpenPluginError) Error() string {
	return errToString("OpenPluginError", e.Err)
}

func (e *NewPluginerError) Unwrap() error {
	return e.Err
}

func (e *NewPluginerError) Error() string {
	return errToString("NewPluginerError", e.Err)
}

func (e *VariableLookupError) Unwrap() error {
	return e.Err
}

func (e *VariableLookupError) Error() string {
	return errToString("VariableLookupError", e.Err)
}

func (e *InvalidConfigError) Unwrap() error {
	return e.Err
}

func (e *InvalidConfigError) Error() string {
	return errToString("InvalidConfigError", e.Err)
}

func (e *NewPluginerNotImplementedError) Unwrap() error {
	return nil
}

func (e *NewPluginerNotImplementedError) Error() string {
	return "NewPluginerNotImplementedError"
}

func (e *AlreadyRegisteredError) Unwrap() error {
	return nil
}

func (e *AlreadyRegisteredError) Error() string {
	return "AlreadyRegisteredError"
}

func (e *NotFoundError) Unwrap() error {
	return nil
}

func (e *NotFoundError) Error() string {
	return "NotFoundError"
}

func (e *NilPluginError) Unwrap() error {
	return nil
}

func (e *NilPluginError) Error() string {
	return "NilPluginError"
}

func errToString(name string, err error) string {
	if err == nil {
		return fmt.Sprintf("%s: unknown", name)
	}

	return fmt.Sprintf("%s: %s", name, err)
}
