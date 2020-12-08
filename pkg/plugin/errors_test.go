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

package plugin_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/sebdah/goldie/v2"
	"github.com/xmidt-org/ears/pkg/plugin"
)

func TestErrorMessage(t *testing.T) {
	testCases := []struct {
		name string
		err  error
	}{
		{name: "Error_Default", err: &plugin.Error{}},

		{
			name: "Error_Err",
			err:  &plugin.Error{Err: fmt.Errorf("wrapped error")},
		},

		{
			name: "Error_Code",
			err:  &plugin.Error{Code: 42},
		},

		{
			name: "Error_All",
			err: &plugin.Error{
				Err:  fmt.Errorf("wrapped error"),
				Code: 42,
			},
		},

		{name: "InvalidConfigError", err: &plugin.InvalidConfigError{}},

		{
			name: "InvalidConfigError_Err",
			err:  &plugin.InvalidConfigError{Err: fmt.Errorf("wrapped error")},
		},

		{name: "NotSupportedError", err: &plugin.NotSupportedError{}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			g := goldie.New(t, goldie.WithTestNameForDir(true))
			g.Assert(t, tc.name, []byte(fmt.Sprint(tc.err)))
			g.Assert(t, tc.name+"_unwrapped", []byte(fmt.Sprint(errors.Unwrap(tc.err))))

		})
	}

}
