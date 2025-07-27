// SPDX-FileCopyrightText: 2025 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package errs_test

import (
	"fmt"

	"github.com/xmidt-org/ears/pkg/errs"
)

type MyError struct {
	Message string
	Code    int
	Err     error
}

func (e *MyError) Error() string {
	return errs.String(
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
