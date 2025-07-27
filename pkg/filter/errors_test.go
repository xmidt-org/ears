// SPDX-FileCopyrightText: 2025 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package filter_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/sebdah/goldie/v2"
	"github.com/xmidt-org/ears/pkg/filter"
)

func TestErrorMessage(t *testing.T) {
	testCases := []struct {
		name string
		err  error
	}{
		{name: "InvalidConfigError", err: &filter.InvalidConfigError{}},

		{
			name: "InvalidConfigError_Err",
			err:  &filter.InvalidConfigError{Err: fmt.Errorf("wrapped error")},
		},

		{name: "InvalidArgumentError", err: &filter.InvalidArgumentError{}},

		{
			name: "InvalidArgumentError_Err",
			err:  &filter.InvalidArgumentError{Err: fmt.Errorf("wrapped error")},
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
