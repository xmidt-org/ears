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

package errors_test

import (
	"fmt"

	"github.com/xmidt-org/ears/pkg/errors"
)

type MyError struct {
	Message string
	Code    int
	Err     error
}

func (e *MyError) Error() string {
	return errors.String(
		"MyError",
		map[string]interface{}{
			"msg":  e.Message,
			"code": e.Code,
		},
		e.Err,
	)
}

func ExampleString() {

	e := &MyError{
		Message: "couldn't compute the meaning of life",
		Code:    -42,
		Err:     fmt.Errorf("observation collapsed the function too hard"),
	}

	fmt.Println(e)
	// Output:  MyError (code=-42 msg=couldn't compute the meaning of life): observation collapsed the function too hard
}
