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

package manager_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/sebdah/goldie/v2"
	"github.com/xmidt-org/ears/pkg/plugin/manager"
)

func TestErrorMessage(t *testing.T) {
	testCases := []struct {
		name string
		err  error
	}{
		{name: "NilPluginError", err: &manager.NilPluginError{}},
		{name: "NotFoundError", err: &manager.NotFoundError{}},
		{name: "AlreadyRegisteredError", err: &manager.AlreadyRegisteredError{}},
		{
			name: "InvalidConfigError_Nil",
			err:  &manager.InvalidConfigError{},
		},
		{
			name: "InvalidConfigError_Err",
			err: &manager.InvalidConfigError{
				Err: fmt.Errorf("wrapped error"),
			},
		},
		{name: "NewPluginerNotImplementedError", err: &manager.NewPluginerNotImplementedError{}},
		{
			name: "VariableLookupError_Nil",
			err:  &manager.VariableLookupError{},
		},
		{
			name: "VariableLookupError_Err",
			err: &manager.VariableLookupError{
				Err: fmt.Errorf("wrapped error"),
			},
		},
		{
			name: "OpenPluginError_Nil",
			err:  &manager.OpenPluginError{},
		},
		{
			name: "OpenPluginError_Err",
			err: &manager.OpenPluginError{
				Err: fmt.Errorf("wrapped error"),
			},
		},
		{
			name: "NewPluginerError_Nil",
			err:  &manager.NewPluginerError{},
		},
		{
			name: "NewPluginerError_Err",
			err: &manager.NewPluginerError{
				Err: fmt.Errorf("wrapped error"),
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			g := goldie.New(t, goldie.WithTestNameForDir(true))
			g.Assert(t, tc.name, []byte(fmt.Sprint(tc.err)))
			g.Assert(t, tc.name+"_unwrapped", []byte(fmt.Sprint(errors.Unwrap(tc.err))))

		})
	}

}
