// SPDX-FileCopyrightText: 2025 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

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
