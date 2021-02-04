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

func TestReceiverWithDefault(t *testing.T) {

	testCases := []struct {
		name     string
		input    debug.Config
		expected debug.Config
	}{
		{
			name:     "empty",
			input:    debug.Config{},
			expected: debug.DefaultConfig,
		},

		{
			name: "intervalMs-zero",
			input: debug.Config{
				IntervalMs: pointer.Int(0),
			},
			expected: debug.Config{
				IntervalMs: pointer.Int(0),
				Rounds:     debug.DefaultConfig.Rounds,
				Payload:    debug.DefaultConfig.Payload,
				MaxHistory: debug.DefaultConfig.MaxHistory,
			},
		},

		{
			name: "intervalMs-negative",
			input: debug.Config{
				IntervalMs: pointer.Int(-3),
			},
			expected: debug.Config{
				IntervalMs: pointer.Int(-3),
				Rounds:     debug.DefaultConfig.Rounds,
				Payload:    debug.DefaultConfig.Payload,
				MaxHistory: debug.DefaultConfig.MaxHistory,
			},
		},

		{
			name: "intervalMs-positive",
			input: debug.Config{
				IntervalMs: pointer.Int(12),
			},
			expected: debug.Config{
				IntervalMs: pointer.Int(12),
				Rounds:     debug.DefaultConfig.Rounds,
				Payload:    debug.DefaultConfig.Payload,
				MaxHistory: debug.DefaultConfig.MaxHistory,
			},
		},

		{
			name: "rounds-negative",
			input: debug.Config{
				Rounds: pointer.Int(-77),
			},
			expected: debug.Config{
				IntervalMs: debug.DefaultConfig.IntervalMs,
				Rounds:     pointer.Int(-77),
				Payload:    debug.DefaultConfig.Payload,
				MaxHistory: debug.DefaultConfig.MaxHistory,
			},
		},

		{
			name: "rounds-zero",
			input: debug.Config{
				Rounds: pointer.Int(0),
			},
			expected: debug.Config{
				IntervalMs: debug.DefaultConfig.IntervalMs,
				Rounds:     pointer.Int(0),
				Payload:    debug.DefaultConfig.Payload,
				MaxHistory: debug.DefaultConfig.MaxHistory,
			},
		},

		{
			name: "rounds-positive",
			input: debug.Config{
				Rounds: pointer.Int(93),
			},
			expected: debug.Config{
				IntervalMs: debug.DefaultConfig.IntervalMs,
				Rounds:     pointer.Int(93),
				Payload:    debug.DefaultConfig.Payload,
				MaxHistory: debug.DefaultConfig.MaxHistory,
			},
		},

		{
			name: "payload-empty",
			input: debug.Config{
				Payload: pointer.String(""),
			},
			expected: debug.Config{
				IntervalMs: debug.DefaultConfig.IntervalMs,
				Rounds:     debug.DefaultConfig.Rounds,
				Payload:    pointer.String(""),
				MaxHistory: debug.DefaultConfig.MaxHistory,
			},
		},

		{
			name: "payload-somevalue",
			input: debug.Config{
				Payload: pointer.String("somevalue"),
			},
			expected: debug.Config{
				IntervalMs: debug.DefaultConfig.IntervalMs,
				Rounds:     debug.DefaultConfig.Rounds,
				Payload:    pointer.String("somevalue"),
				MaxHistory: debug.DefaultConfig.MaxHistory,
			},
		},

		{
			name: "maxHistory-negative",
			input: debug.Config{
				MaxHistory: pointer.Int(-8),
			},
			expected: debug.Config{
				IntervalMs: debug.DefaultConfig.IntervalMs,
				Rounds:     debug.DefaultConfig.Rounds,
				Payload:    debug.DefaultConfig.Payload,
				MaxHistory: pointer.Int(-8),
			},
		},

		{
			name: "maxHistory-zero",
			input: debug.Config{
				MaxHistory: pointer.Int(0),
			},
			expected: debug.Config{
				IntervalMs: debug.DefaultConfig.IntervalMs,
				Rounds:     debug.DefaultConfig.Rounds,
				Payload:    debug.DefaultConfig.Payload,
				MaxHistory: pointer.Int(0),
			},
		},

		{
			name: "maxHistory-positive",
			input: debug.Config{
				MaxHistory: pointer.Int(26),
			},
			expected: debug.Config{
				IntervalMs: debug.DefaultConfig.IntervalMs,
				Rounds:     debug.DefaultConfig.Rounds,
				Payload:    debug.DefaultConfig.Payload,
				MaxHistory: pointer.Int(26),
			},
		},

		{
			name: "all",
			input: debug.Config{
				IntervalMs: pointer.Int(54),
				Rounds:     pointer.Int(47),
				Payload:    "something expected",
				MaxHistory: pointer.Int(2836),
			},
			expected: debug.Config{
				IntervalMs: pointer.Int(54),
				Rounds:     pointer.Int(47),
				Payload:    "something expected",
				MaxHistory: pointer.Int(2836),
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

func TestReceiverValidation(t *testing.T) {
	testCases := []struct {
		name  string
		input debug.Config
		valid bool
	}{
		{
			name:  "empty",
			input: debug.Config{},
			valid: false,
		},

		{
			name:  "default",
			input: debug.DefaultConfig,
			valid: true,
		},
		{
			name:  "with-defaults",
			input: (&debug.Config{}).WithDefaults(),
			valid: true,
		},

		{
			name: "full",
			input: debug.Config{
				IntervalMs: pointer.Int(654),
				Rounds:     pointer.Int(4007),
				Payload:    "payload here",
				MaxHistory: pointer.Int(836),
			},
			valid: true,
		},

		{
			name: "intervalMs-unset",
			input: debug.Config{
				IntervalMs: nil,
				Rounds:     debug.DefaultConfig.Rounds,
				Payload:    debug.DefaultConfig.Payload,
				MaxHistory: debug.DefaultConfig.MaxHistory,
			},
			valid: false,
		},

		{
			name: "intervalMs-zero",
			input: debug.Config{
				IntervalMs: pointer.Int(0),
				Rounds:     debug.DefaultConfig.Rounds,
				Payload:    debug.DefaultConfig.Payload,
				MaxHistory: debug.DefaultConfig.MaxHistory,
			},
			valid: false,
		},

		{
			name: "intervalMs-positive",
			input: debug.Config{
				IntervalMs: pointer.Int(1),
				Rounds:     debug.DefaultConfig.Rounds,
				Payload:    debug.DefaultConfig.Payload,
				MaxHistory: debug.DefaultConfig.MaxHistory,
			},
			valid: true,
		},

		{
			name: "intervalMs-negative",
			input: debug.Config{
				IntervalMs: pointer.Int(-10),
				Rounds:     debug.DefaultConfig.Rounds,
				Payload:    debug.DefaultConfig.Payload,
				MaxHistory: debug.DefaultConfig.MaxHistory,
			},
			valid: false,
		},

		{
			name: "rounds-unset",
			input: debug.Config{
				IntervalMs: debug.DefaultConfig.IntervalMs,
				Rounds:     nil,
				Payload:    debug.DefaultConfig.Payload,
				MaxHistory: debug.DefaultConfig.MaxHistory,
			},
			valid: false,
		},

		{
			name: "rounds-zero",
			input: debug.Config{
				IntervalMs: debug.DefaultConfig.IntervalMs,
				Rounds:     pointer.Int(0),
				Payload:    debug.DefaultConfig.Payload,
				MaxHistory: debug.DefaultConfig.MaxHistory,
			},
			valid: true,
		},

		{
			name: "rounds-negative",
			input: debug.Config{
				IntervalMs: debug.DefaultConfig.IntervalMs,
				Rounds:     pointer.Int(-1),
				Payload:    debug.DefaultConfig.Payload,
				MaxHistory: debug.DefaultConfig.MaxHistory,
			},
			valid: true,
		},

		{
			name: "rounds-positive",
			input: debug.Config{
				IntervalMs: debug.DefaultConfig.IntervalMs,
				Rounds:     pointer.Int(1),
				Payload:    debug.DefaultConfig.Payload,
				MaxHistory: debug.DefaultConfig.MaxHistory,
			},
			valid: true,
		},

		{
			name: "maxhistory-unset",
			input: debug.Config{
				IntervalMs: debug.DefaultConfig.IntervalMs,
				Rounds:     debug.DefaultConfig.Rounds,
				Payload:    debug.DefaultConfig.Payload,
				MaxHistory: nil,
			},
			valid: false,
		},

		{
			name: "maxhistory-zero",
			input: debug.Config{
				IntervalMs: debug.DefaultConfig.IntervalMs,
				Rounds:     debug.DefaultConfig.Rounds,
				Payload:    debug.DefaultConfig.Payload,
				MaxHistory: pointer.Int(0),
			},
			valid: true,
		},

		{
			name: "maxhistory-positive",
			input: debug.Config{
				IntervalMs: debug.DefaultConfig.IntervalMs,
				Rounds:     debug.DefaultConfig.Rounds,
				Payload:    debug.DefaultConfig.Payload,
				MaxHistory: pointer.Int(463),
			},
			valid: true,
		},

		{
			name: "maxhistory-negative",
			input: debug.Config{
				IntervalMs: debug.DefaultConfig.IntervalMs,
				Rounds:     debug.DefaultConfig.Rounds,
				Payload:    debug.DefaultConfig.Payload,
				MaxHistory: pointer.Int(-34),
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

func TestReceiverSerialization(t *testing.T) {
	testCases := []struct {
		name   string
		config debug.Config
	}{
		{
			name:   "empty",
			config: debug.Config{},
		},

		{
			name:   "default",
			config: (&debug.Config{}).WithDefaults(),
		},

		{
			name:   "interval-zero",
			config: debug.Config{IntervalMs: pointer.Int(0)},
		},

		{
			name:   "interval-negative",
			config: debug.Config{IntervalMs: pointer.Int(-882)},
		},

		{
			name:   "interval-positive",
			config: debug.Config{IntervalMs: pointer.Int(1000)},
		},

		{
			name:   "rounds-zero",
			config: debug.Config{Rounds: pointer.Int(0)},
		},

		{
			name:   "rounds-negative",
			config: debug.Config{Rounds: pointer.Int(-234)},
		},

		{
			name:   "rounds-positive",
			config: debug.Config{Rounds: pointer.Int(2372)},
		},

		{
			name:   "payload-empty",
			config: debug.Config{Payload: pointer.String("")},
		},

		{
			name:   "payload-withvalue",
			config: debug.Config{Payload: pointer.String("some value")},
		},

		{
			name:   "maxhistory-zero",
			config: debug.Config{MaxHistory: pointer.Int(0)},
		},

		{
			name:   "maxhistory-positive",
			config: debug.Config{MaxHistory: pointer.Int(2347)},
		},

		{
			name:   "maxhistory-negative",
			config: debug.Config{MaxHistory: pointer.Int(-4)},
		},

		{
			name:   "maxhistory-maxint",
			config: debug.Config{MaxHistory: pointer.Int(9223372036854775807)},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			g := goldie.New(t, goldie.WithTestNameForDir(true))
			g.AssertJson(t, tc.name, tc.config)
		})
	}

}
