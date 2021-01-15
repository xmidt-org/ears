// Copyright 2021 Comcast Cable Communications Management, LLC
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

package config_test

import (
	"testing"

	. "github.com/onsi/gomega"
	"github.com/xmidt-org/ears/pkg/config"
	"github.com/xmidt-org/ears/pkg/errs"
)

type testConfig1 struct {
	Name   string
	IntVal int
}

type testConfig2 struct {
	Name   string
	IntVal int
}

func TestNewConfig(t *testing.T) {
	testCases := []struct {
		name           string
		data           interface{}
		config         interface{}
		expectedConfig interface{}
		expectedErr    error
	}{
		{
			name:           "nil-data",
			data:           nil,
			config:         &testConfig1{},
			expectedConfig: nil,
			expectedErr:    &config.InvalidArgumentError{},
		},
		{
			name:           "nil-config",
			data:           "data",
			config:         nil,
			expectedConfig: nil,
			expectedErr:    &config.InvalidArgumentError{},
		},
		{
			name:           "parse-str-error",
			data:           "1234",
			config:         &testConfig1{},
			expectedConfig: nil,
			expectedErr:    &config.DataParseError{},
		},
		{
			name:           "parse-rune-error",
			data:           []rune("1234"),
			config:         &testConfig1{},
			expectedConfig: nil,
			expectedErr:    &config.DataParseError{},
		},
		{
			name:           "incompatible-types",
			data:           1234,
			config:         &testConfig1{},
			expectedConfig: &testConfig1{},
			expectedErr:    &config.InvalidArgumentError{},
		},
		{
			name:           "yaml-str-to-obj",
			data:           `name: myname`,
			config:         &testConfig1{},
			expectedConfig: &testConfig1{Name: "myname"},
			expectedErr:    nil,
		},
		{
			name:           "yaml-bytes-to-obj",
			data:           []byte(`name: myname`),
			config:         &testConfig1{},
			expectedConfig: &testConfig1{Name: "myname"},
			expectedErr:    nil,
		},
		{
			name:           "yaml-rune-to-obj",
			data:           []rune("name: 日本語"),
			config:         &testConfig1{},
			expectedConfig: &testConfig1{Name: "日本語"},
			expectedErr:    nil,
		},
		{
			name:           "json-str-to-obj",
			data:           `{"name": "myname"}`,
			config:         &testConfig1{},
			expectedConfig: &testConfig1{Name: "myname"},
			expectedErr:    nil,
		},
		{
			name:           "json-bytes-to-obj",
			data:           []byte(`{"name": "myname"}`),
			config:         &testConfig1{},
			expectedConfig: &testConfig1{Name: "myname"},
			expectedErr:    nil,
		},
		{
			name:           "json-rune-to-obj",
			data:           []rune(`{"name": "myname"}`),
			config:         &testConfig1{},
			expectedConfig: &testConfig1{Name: "myname"},
			expectedErr:    nil,
		},
		{
			name:           "object-to-object",
			data:           testConfig1{Name: "hello", IntVal: 9},
			config:         testConfig1{},
			expectedConfig: nil,
			expectedErr:    &config.InvalidArgumentError{}, // Target not pointer
		},
		{
			name:           "object-to-ptr",
			data:           testConfig1{Name: "hello", IntVal: 9},
			config:         &testConfig1{},
			expectedConfig: &testConfig1{Name: "hello", IntVal: 9},
			expectedErr:    nil,
		},
		{
			name:           "ptr-to-object",
			data:           &testConfig1{Name: "hello", IntVal: 9},
			config:         testConfig1{},
			expectedConfig: nil,
			expectedErr:    &config.InvalidArgumentError{}, // Target not pointer
		},
		{
			name:           "ptr-to-ptr",
			data:           &testConfig1{Name: "hello", IntVal: 9},
			config:         &testConfig1{},
			expectedConfig: &testConfig1{Name: "hello", IntVal: 9},
			expectedErr:    nil,
		},
		{
			name:           "ptrcfg1-to-ptrcfg2",
			data:           &testConfig2{Name: "hello", IntVal: 9},
			config:         &testConfig1{},
			expectedConfig: nil,
			expectedErr:    &config.InvalidArgumentError{}, // Target not pointer
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			a := NewWithT(t)

			err := config.NewConfig(tc.data, tc.config)

			if tc.expectedErr != nil {
				a.Expect(errs.Type(err)).To(Equal(errs.Type(tc.expectedErr)))
			} else {
				a.Expect(err).To(BeNil())
				a.Expect(tc.config).To(Equal(tc.expectedConfig))
			}
		})
	}

}
