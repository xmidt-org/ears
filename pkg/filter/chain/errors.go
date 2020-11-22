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

package chain

import "fmt"

func (e *InvalidArgumentError) Unwrap() error {
	return e.Err
}

func (e *InvalidArgumentError) Error() string {
	return errToString("InvalidArgumentError", e.Err)
}

func errToString(name string, err error) string {
	if err == nil {
		return name
	}

	return fmt.Sprintf("%s: %s", name, err)
}
