// SPDX-FileCopyrightText: 2025 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package config_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/sebdah/goldie/v2"
	"github.com/xmidt-org/ears/pkg/config"
)

func TestErrorMessage(t *testing.T) {
	testCases := []struct {
		name string
		err  error
	}{
		{name: "Error-empty", err: &config.Error{}},
		{
			name: "Error-err",
			err: &config.Error{
				Err: fmt.Errorf("wrapped error"),
			},
		},
		{name: "InvalidArgumentError-empty", err: &config.InvalidArgumentError{}},
		{
			name: "InvalidArgumentError-err",
			err: &config.InvalidArgumentError{
				Err: fmt.Errorf("wrapped error"),
			},
		},
		{name: "DataParseError-empty", err: &config.DataParseError{}},
		{
			name: "DataParseError-err",
			err: &config.DataParseError{
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
