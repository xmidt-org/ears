// SPDX-FileCopyrightText: 2025 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package errs_test

import (
	"fmt"
	"testing"

	"github.com/sebdah/goldie/v2"
	"github.com/xmidt-org/ears/pkg/errs"
)

func TestString(t *testing.T) {
	testCases := []struct {
		id      string
		name    string
		values  map[string]interface{}
		wrapped error
	}{

		{
			id:      "empty-all",
			name:    "",
			values:  nil,
			wrapped: nil,
		},

		{
			id:      "basic-name",
			name:    "basic-name",
			values:  nil,
			wrapped: nil,
		},

		{
			id:      "basic-values",
			name:    "",
			values:  map[string]interface{}{"first": true, "second": 2, "third": "third", "forth": []int{1, 2, 3, 4}},
			wrapped: nil,
		},

		{
			id:      "basic-wrapped",
			name:    "",
			values:  nil,
			wrapped: fmt.Errorf("wrapped error"),
		},

		{
			id:      "name-values",
			name:    "name-values",
			values:  map[string]interface{}{"first": true, "second": 2, "third": "third", "forth": []int{1, 2, 3, 4}},
			wrapped: nil,
		},

		{
			id:      "name-wrapped",
			name:    "name-wrapped",
			values:  nil,
			wrapped: fmt.Errorf("wrapped error"),
		},

		{
			id:      "name-values-wrapped",
			name:    "name-values",
			values:  map[string]interface{}{"first": true, "second": 2, "third": "third", "forth": []int{1, 2, 3, 4}},
			wrapped: fmt.Errorf("wrapped error"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.id, func(t *testing.T) {
			g := goldie.New(t, goldie.WithTestNameForDir(true))
			g.Assert(t, tc.id, []byte(errs.String(tc.name, tc.values, tc.wrapped)))
		})
	}

}
