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
	"github.com/xmidt-org/ears/internal/pkg/plugin"
)

func TestErrorMessage(t *testing.T) {
	testCases := []struct {
		name string
		err  error
	}{
		{name: "OptionError-empty", err: &plugin.OptionError{}},
		{
			name: "OptionError-message",
			err: &plugin.OptionError{
				Message: "message goes here",
			},
		},
		{
			name: "OptionError-err",
			err: &plugin.OptionError{
				Err: fmt.Errorf("wrapped error"),
			},
		},
		{
			name: "OptionError-message-err",
			err: &plugin.OptionError{
				Message: "message goes here",
				Err:     fmt.Errorf("wrapped error"),
			},
		},

		{name: "RegistrationError-empty", err: &plugin.RegistrationError{}},

		{
			name: "RegistrationError-error",
			err: &plugin.RegistrationError{
				Err: fmt.Errorf("wrapped error"),
			},
		},

		{
			name: "RegistrationError-message",
			err: &plugin.RegistrationError{
				Message: "message goes here",
			},
		},

		{
			name: "RegistrationError-message-error",
			err: &plugin.RegistrationError{
				Message: "message goes here",
				Err:     fmt.Errorf("wrapped error"),
			},
		},

		{
			name: "RegistrationError-plugin",
			err: &plugin.RegistrationError{
				Plugin: "plugin goes here",
			},
		},

		{
			name: "RegistrationError-plugin-error",
			err: &plugin.RegistrationError{
				Plugin: "plugin goes here",
				Err:    fmt.Errorf("wrapped error"),
			},
		},

		{
			name: "RegistrationError-name",
			err: &plugin.RegistrationError{
				Name: "name goes here",
			},
		},

		{
			name: "RegistrationError-name-error",
			err: &plugin.RegistrationError{
				Name: "name goes here",
				Err:  fmt.Errorf("wrapped error"),
			},
		},

		{
			name: "RegistrationError-message-plugin-name-error",
			err: &plugin.RegistrationError{
				Message: "message goes here",
				Name:    "name goes here",
				Plugin:  "plugin goes here",
				Err:     fmt.Errorf("wrapped error"),
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
