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

package debug_test

import (
	"testing"

	"github.com/xmidt-org/ears/pkg/plugins/debug"
	"github.com/xorcare/pointer"

	"github.com/sebdah/goldie/v2"

	. "github.com/onsi/gomega"
)

func TestSenderWithDefault(t *testing.T) {

	testCases := []struct {
		name     string
		input    debug.SenderConfig
		expected debug.SenderConfig
	}{
		{
			name:     "empty",
			input:    debug.SenderConfig{},
			expected: debug.DefaultSenderConfig,
		},
		{
			name: "maxHistory",
			input: debug.SenderConfig{
				MaxHistory: pointer.Int(12),
			},
			expected: debug.SenderConfig{
				Destination: debug.DefaultSenderConfig.Destination,
				MaxHistory:  pointer.Int(12),
				Writer:      debug.DefaultSenderConfig.Writer,
			},
		},

		{
			name: "destination",
			input: debug.SenderConfig{
				Destination: debug.DestinationStderr,
			},
			expected: debug.SenderConfig{
				Destination: debug.DestinationStderr,
				MaxHistory:  debug.DefaultSenderConfig.MaxHistory,
				Writer:      debug.DefaultSenderConfig.Writer,
			},
		},

		{
			name: "writer",
			input: debug.SenderConfig{
				Destination: debug.DestinationCustom,
				Writer:      &debug.SendSlice{},
			},
			expected: debug.SenderConfig{
				Destination: debug.DestinationCustom,
				MaxHistory:  debug.DefaultSenderConfig.MaxHistory,
				Writer:      &debug.SendSlice{},
			},
		},

		{
			name: "all",
			input: debug.SenderConfig{
				Destination: debug.DestinationCustom,
				MaxHistory:  pointer.Int(19),
				Writer:      &debug.SendStdout{},
			},
			expected: debug.SenderConfig{
				Destination: debug.DestinationCustom,
				MaxHistory:  pointer.Int(19),
				Writer:      &debug.SendStdout{},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			a := NewWithT(t)

			d := tc.input.WithDefaults()
			a.Expect(d).To(Equal(tc.expected))

		})
	}

}

func TestSenderValidation(t *testing.T) {
	testCases := []struct {
		name  string
		input debug.SenderConfig
		valid bool
	}{
		{
			name:  "default",
			input: debug.DefaultSenderConfig,
			valid: true,
		},
		{
			name:  "with-defaults",
			input: (&debug.SenderConfig{}).WithDefaults(),
			valid: true,
		},

		{
			name:  "empty",
			input: debug.SenderConfig{},
			valid: false,
		},

		{
			name: "full",
			input: debug.SenderConfig{
				Destination: debug.DefaultSenderConfig.Destination,
				MaxHistory:  debug.DefaultSenderConfig.MaxHistory,
				Writer:      debug.DefaultSenderConfig.Writer,
			},
			valid: true,
		},

		{
			name: "destination-unset",
			input: debug.SenderConfig{
				MaxHistory: debug.DefaultSenderConfig.MaxHistory,
				Writer:     debug.DefaultSenderConfig.Writer,
			},
			valid: false,
		},

		{
			name: "destination-unknown",
			input: debug.SenderConfig{
				Destination: debug.DestinationUnknown,
				MaxHistory:  debug.DefaultSenderConfig.MaxHistory,
				Writer:      debug.DefaultSenderConfig.Writer,
			},
			valid: false,
		},

		{
			name: "destination-23",
			input: debug.SenderConfig{
				Destination: 23,
				MaxHistory:  debug.DefaultSenderConfig.MaxHistory,
				Writer:      debug.DefaultSenderConfig.Writer,
			},
			valid: false,
		},

		{
			name: "maxHistory-negative",
			input: debug.SenderConfig{
				Destination: debug.DefaultSenderConfig.Destination,
				MaxHistory:  pointer.Int(-1),
				Writer:      debug.DefaultSenderConfig.Writer,
			},
			valid: false,
		},

		{
			name: "maxHistory-nil",
			input: debug.SenderConfig{
				Destination: debug.DefaultSenderConfig.Destination,
				MaxHistory:  nil,
				Writer:      debug.DefaultSenderConfig.Writer,
			},
			valid: false,
		},

		{
			name: "maxHistory-zero",
			input: debug.SenderConfig{
				Destination: debug.DefaultSenderConfig.Destination,
				MaxHistory:  pointer.Int(0),
				Writer:      debug.DefaultSenderConfig.Writer,
			},
			valid: true,
		},

		{
			name: "writer-nil",
			input: debug.SenderConfig{
				Destination: debug.DefaultSenderConfig.Destination,
				MaxHistory:  debug.DefaultSenderConfig.MaxHistory,
				Writer:      nil,
			},
			valid: true,
		},

		{
			name: "writer-custom",
			input: debug.SenderConfig{
				Destination: debug.DestinationCustom,
				MaxHistory:  debug.DefaultSenderConfig.MaxHistory,
				Writer:      &debug.SendStderr{},
			},
			valid: true,
		},

		{
			name: "writer-custom-but-nil",
			input: debug.SenderConfig{
				Destination: debug.DestinationCustom,
				MaxHistory:  debug.DefaultSenderConfig.MaxHistory,
				Writer:      nil,
			},
			valid: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			a := NewWithT(t)

			err := (&tc.input).Validate()
			if tc.valid {
				a.Expect(err).To(BeNil())
			} else {
				a.Expect(err).ToNot(BeNil())
			}

		})

	}

}

func TestSenderSerialization(t *testing.T) {
	testCases := []struct {
		name   string
		config debug.SenderConfig
	}{
		{
			name:   "empty",
			config: debug.SenderConfig{},
		},

		{
			name:   "default",
			config: (&debug.SenderConfig{}).WithDefaults(),
		},

		{
			name:   "devnull",
			config: debug.SenderConfig{Destination: debug.DestinationDevNull},
		},

		{
			name:   "stdout",
			config: debug.SenderConfig{Destination: debug.DestinationStdout},
		},

		{
			name:   "stderr",
			config: debug.SenderConfig{Destination: debug.DestinationStderr},
		},

		{
			name:   "custom",
			config: debug.SenderConfig{Destination: debug.DestinationCustom},
		},

		{
			name:   "maxhistory-nil",
			config: debug.SenderConfig{MaxHistory: nil},
		},

		{
			name:   "maxhistory-zero",
			config: debug.SenderConfig{MaxHistory: pointer.Int(0)},
		},

		{
			name:   "maxhistory-negative",
			config: debug.SenderConfig{MaxHistory: pointer.Int(-4)},
		},

		{
			name:   "maxhistory-positive",
			config: debug.SenderConfig{MaxHistory: pointer.Int(99999999999)},
		},

		{
			name:   "maxhistory-maxint",
			config: debug.SenderConfig{MaxHistory: pointer.Int(9223372036854775807)},
		},

		{
			name:   "maxhistory-meaningoflife",
			config: debug.SenderConfig{MaxHistory: pointer.Int(42)},
		},

		{
			name: "writer-custom",
			config: debug.SenderConfig{
				Destination: debug.DestinationCustom,
				Writer:      &debug.SendSlice{},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			g := goldie.New(t, goldie.WithTestNameForDir(true))
			g.AssertJson(t, tc.name, tc.config)
		})
	}

}
