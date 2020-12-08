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

package app

import (
	"errors"
	"testing"
)

type MyTestError struct {
}

func (e *MyTestError) Error() string {
	return "A unit test error"
}

func TestErrors(t *testing.T) {

	err := &Error{
		Err: &MyTestError{},
	}
	var myTestErr *MyTestError
	if !errors.As(err, &myTestErr) {
		t.Errorf("Error type check failed")
	}

	err2 := &Error{
		Err: &MyTestError{},
	}
	if !errors.Is(err, err2) {
		t.Error("Error comparison failed")
	}

	err3 := &InvalidOptionError{
		Err: &MyTestError{},
	}
	if errors.Is(err, err3) {
		t.Error("Error comparison should fail but not failing")
	}

	if !errors.As(err3, &myTestErr) {
		t.Errorf("Error type check failed")
	}

	err4 := &InvalidOptionError{
		Err: &MyTestError{},
	}
	if !errors.Is(err3, err4) {
		t.Error("Error comparison failed")
	}

}
