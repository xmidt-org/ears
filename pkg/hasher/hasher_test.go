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

package hasher_test

import (
	"testing"

	"github.com/sebdah/goldie/v2"
	"github.com/xmidt-org/ears/pkg/hasher"
)

func TestString(t *testing.T) {
	testCases := []struct {
		name  string
		value string
	}{
		{name: "empty", value: ""},
		{name: "alpha", value: "abcdefghijklmnopqrstuvwxyz"},
		{name: "numeric", value: "1234567890"},
		{name: "bytes", value: string([]byte{1, 2, 3, 4})},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			g := goldie.New(t, goldie.WithTestNameForDir(true))

			g.Assert(t, tc.name, []byte(hasher.String(tc.value)))
		})
	}
}

func TestHash(t *testing.T) {
	testCases := []struct {
		name  string
		value interface{}
	}{
		{name: "bool", value: true},
		{name: "[]bool", value: []bool{true, false}},
		{name: "string", value: "abcdefgh"},
		{name: "string_empty", value: ""},
		{name: "[]string", value: []string{"1234567890", "0987654321"}},
		{name: "int", value: 42},
		{name: "[]int", value: []int64{-34, 0, 299792458}},
		{name: "float", value: 3.14},
		{name: "[]float", value: []float32{3.14, 2.718}},
		{name: "[]byte", value: []byte{1, 2, 3, 4}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			g := goldie.New(t, goldie.WithTestNameForDir(true))

			g.Assert(t, tc.name, []byte(hasher.Hash(tc.value)))
		})
	}
}
