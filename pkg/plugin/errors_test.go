// SPDX-FileCopyrightText: 2025 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

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
