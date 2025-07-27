// SPDX-FileCopyrightText: 2025 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package sender_test

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
		{name: "InvalidConfigError", err: &plugin.InvalidConfigError{}},

		{
			name: "InvalidConfigError_Err",
			err:  &plugin.InvalidConfigError{Err: fmt.Errorf("wrapped error")},
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
