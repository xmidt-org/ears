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
