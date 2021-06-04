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
		input    debug.ReceiverConfig
		expected debug.ReceiverConfig
	}{
		{
			name:     "empty",
			input:    debug.ReceiverConfig{},
			expected: debug.DefaultReceiverConfig,
		},

		{
			name: "intervalMs-zero",
			input: debug.ReceiverConfig{
				IntervalMs: pointer.Int(0),
			},
			expected: debug.ReceiverConfig{
				IntervalMs: pointer.Int(0),
				Rounds:     debug.DefaultReceiverConfig.Rounds,
				Payload:    debug.DefaultReceiverConfig.Payload,
				MaxHistory: debug.DefaultReceiverConfig.MaxHistory,
				Trace:      debug.DefaultReceiverConfig.Trace,
			},
		},

		{
			name: "intervalMs-negative",
			input: debug.ReceiverConfig{
				IntervalMs: pointer.Int(-3),
			},
			expected: debug.ReceiverConfig{
				IntervalMs: pointer.Int(-3),
				Rounds:     debug.DefaultReceiverConfig.Rounds,
				Payload:    debug.DefaultReceiverConfig.Payload,
				MaxHistory: debug.DefaultReceiverConfig.MaxHistory,
				Trace:      debug.DefaultReceiverConfig.Trace,
			},
		},

		{
			name: "intervalMs-positive",
			input: debug.ReceiverConfig{
				IntervalMs: pointer.Int(12),
			},
			expected: debug.ReceiverConfig{
				IntervalMs: pointer.Int(12),
				Rounds:     debug.DefaultReceiverConfig.Rounds,
				Payload:    debug.DefaultReceiverConfig.Payload,
				MaxHistory: debug.DefaultReceiverConfig.MaxHistory,
				Trace:      debug.DefaultReceiverConfig.Trace,
			},
		},

		{
			name: "rounds-negative",
			input: debug.ReceiverConfig{
				Rounds: pointer.Int(-77),
			},
			expected: debug.ReceiverConfig{
				IntervalMs: debug.DefaultReceiverConfig.IntervalMs,
				Rounds:     pointer.Int(-77),
				Payload:    debug.DefaultReceiverConfig.Payload,
				MaxHistory: debug.DefaultReceiverConfig.MaxHistory,
				Trace:      debug.DefaultReceiverConfig.Trace,
			},
		},

		{
			name: "rounds-zero",
			input: debug.ReceiverConfig{
				Rounds: pointer.Int(0),
			},
			expected: debug.ReceiverConfig{
				IntervalMs: debug.DefaultReceiverConfig.IntervalMs,
				Rounds:     pointer.Int(0),
				Payload:    debug.DefaultReceiverConfig.Payload,
				MaxHistory: debug.DefaultReceiverConfig.MaxHistory,
				Trace:      debug.DefaultReceiverConfig.Trace,
			},
		},

		{
			name: "rounds-positive",
			input: debug.ReceiverConfig{
				Rounds: pointer.Int(93),
			},
			expected: debug.ReceiverConfig{
				IntervalMs: debug.DefaultReceiverConfig.IntervalMs,
				Rounds:     pointer.Int(93),
				Payload:    debug.DefaultReceiverConfig.Payload,
				MaxHistory: debug.DefaultReceiverConfig.MaxHistory,
				Trace:      debug.DefaultReceiverConfig.Trace,
			},
		},

		{
			name: "payload-empty",
			input: debug.ReceiverConfig{
				Payload: pointer.String(""),
			},
			expected: debug.ReceiverConfig{
				IntervalMs: debug.DefaultReceiverConfig.IntervalMs,
				Rounds:     debug.DefaultReceiverConfig.Rounds,
				Payload:    pointer.String(""),
				MaxHistory: debug.DefaultReceiverConfig.MaxHistory,
				Trace:      debug.DefaultReceiverConfig.Trace,
			},
		},

		{
			name: "payload-somevalue",
			input: debug.ReceiverConfig{
				Payload: pointer.String("somevalue"),
			},
			expected: debug.ReceiverConfig{
				IntervalMs: debug.DefaultReceiverConfig.IntervalMs,
				Rounds:     debug.DefaultReceiverConfig.Rounds,
				Payload:    pointer.String("somevalue"),
				MaxHistory: debug.DefaultReceiverConfig.MaxHistory,
				Trace:      debug.DefaultReceiverConfig.Trace,
			},
		},

		{
			name: "maxHistory-negative",
			input: debug.ReceiverConfig{
				MaxHistory: pointer.Int(-8),
			},
			expected: debug.ReceiverConfig{
				IntervalMs: debug.DefaultReceiverConfig.IntervalMs,
				Rounds:     debug.DefaultReceiverConfig.Rounds,
				Payload:    debug.DefaultReceiverConfig.Payload,
				MaxHistory: pointer.Int(-8),
				Trace:      debug.DefaultReceiverConfig.Trace,
			},
		},

		{
			name: "maxHistory-zero",
			input: debug.ReceiverConfig{
				MaxHistory: pointer.Int(0),
			},
			expected: debug.ReceiverConfig{
				IntervalMs: debug.DefaultReceiverConfig.IntervalMs,
				Rounds:     debug.DefaultReceiverConfig.Rounds,
				Payload:    debug.DefaultReceiverConfig.Payload,
				MaxHistory: pointer.Int(0),
				Trace:      debug.DefaultReceiverConfig.Trace,
			},
		},

		{
			name: "maxHistory-positive",
			input: debug.ReceiverConfig{
				MaxHistory: pointer.Int(26),
			},
			expected: debug.ReceiverConfig{
				IntervalMs: debug.DefaultReceiverConfig.IntervalMs,
				Rounds:     debug.DefaultReceiverConfig.Rounds,
				Payload:    debug.DefaultReceiverConfig.Payload,
				MaxHistory: pointer.Int(26),
				Trace:      debug.DefaultReceiverConfig.Trace,
			},
		},

		{
			name: "all",
			input: debug.ReceiverConfig{
				IntervalMs: pointer.Int(54),
				Rounds:     pointer.Int(47),
				Payload:    "something expected",
				MaxHistory: pointer.Int(2836),
				Trace:      debug.DefaultReceiverConfig.Trace,
			},
			expected: debug.ReceiverConfig{
				IntervalMs: pointer.Int(54),
				Rounds:     pointer.Int(47),
				Payload:    "something expected",
				MaxHistory: pointer.Int(2836),
				Trace:      debug.DefaultReceiverConfig.Trace,
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
		input debug.ReceiverConfig
		valid bool
	}{
		{
			name:  "empty",
			input: debug.ReceiverConfig{},
			valid: false,
		},

		{
			name:  "default",
			input: debug.DefaultReceiverConfig,
			valid: true,
		},
		{
			name:  "with-defaults",
			input: (&debug.ReceiverConfig{}).WithDefaults(),
			valid: true,
		},

		{
			name: "full",
			input: debug.ReceiverConfig{
				IntervalMs: pointer.Int(654),
				Rounds:     pointer.Int(4007),
				Payload:    "payload here",
				MaxHistory: pointer.Int(836),
			},
			valid: true,
		},

		{
			name: "intervalMs-unset",
			input: debug.ReceiverConfig{
				IntervalMs: nil,
				Rounds:     debug.DefaultReceiverConfig.Rounds,
				Payload:    debug.DefaultReceiverConfig.Payload,
				MaxHistory: debug.DefaultReceiverConfig.MaxHistory,
			},
			valid: false,
		},

		{
			name: "intervalMs-zero",
			input: debug.ReceiverConfig{
				IntervalMs: pointer.Int(0),
				Rounds:     debug.DefaultReceiverConfig.Rounds,
				Payload:    debug.DefaultReceiverConfig.Payload,
				MaxHistory: debug.DefaultReceiverConfig.MaxHistory,
			},
			valid: false,
		},

		{
			name: "intervalMs-positive",
			input: debug.ReceiverConfig{
				IntervalMs: pointer.Int(1),
				Rounds:     debug.DefaultReceiverConfig.Rounds,
				Payload:    debug.DefaultReceiverConfig.Payload,
				MaxHistory: debug.DefaultReceiverConfig.MaxHistory,
			},
			valid: true,
		},

		{
			name: "intervalMs-negative",
			input: debug.ReceiverConfig{
				IntervalMs: pointer.Int(-10),
				Rounds:     debug.DefaultReceiverConfig.Rounds,
				Payload:    debug.DefaultReceiverConfig.Payload,
				MaxHistory: debug.DefaultReceiverConfig.MaxHistory,
			},
			valid: false,
		},

		{
			name: "rounds-unset",
			input: debug.ReceiverConfig{
				IntervalMs: debug.DefaultReceiverConfig.IntervalMs,
				Rounds:     nil,
				Payload:    debug.DefaultReceiverConfig.Payload,
				MaxHistory: debug.DefaultReceiverConfig.MaxHistory,
			},
			valid: false,
		},

		{
			name: "rounds-zero",
			input: debug.ReceiverConfig{
				IntervalMs: debug.DefaultReceiverConfig.IntervalMs,
				Rounds:     pointer.Int(0),
				Payload:    debug.DefaultReceiverConfig.Payload,
				MaxHistory: debug.DefaultReceiverConfig.MaxHistory,
			},
			valid: true,
		},

		{
			name: "rounds-negative",
			input: debug.ReceiverConfig{
				IntervalMs: debug.DefaultReceiverConfig.IntervalMs,
				Rounds:     pointer.Int(-1),
				Payload:    debug.DefaultReceiverConfig.Payload,
				MaxHistory: debug.DefaultReceiverConfig.MaxHistory,
			},
			valid: true,
		},

		{
			name: "rounds-positive",
			input: debug.ReceiverConfig{
				IntervalMs: debug.DefaultReceiverConfig.IntervalMs,
				Rounds:     pointer.Int(1),
				Payload:    debug.DefaultReceiverConfig.Payload,
				MaxHistory: debug.DefaultReceiverConfig.MaxHistory,
			},
			valid: true,
		},

		{
			name: "maxhistory-unset",
			input: debug.ReceiverConfig{
				IntervalMs: debug.DefaultReceiverConfig.IntervalMs,
				Rounds:     debug.DefaultReceiverConfig.Rounds,
				Payload:    debug.DefaultReceiverConfig.Payload,
				MaxHistory: nil,
			},
			valid: false,
		},

		{
			name: "maxhistory-zero",
			input: debug.ReceiverConfig{
				IntervalMs: debug.DefaultReceiverConfig.IntervalMs,
				Rounds:     debug.DefaultReceiverConfig.Rounds,
				Payload:    debug.DefaultReceiverConfig.Payload,
				MaxHistory: pointer.Int(0),
			},
			valid: true,
		},

		{
			name: "maxhistory-positive",
			input: debug.ReceiverConfig{
				IntervalMs: debug.DefaultReceiverConfig.IntervalMs,
				Rounds:     debug.DefaultReceiverConfig.Rounds,
				Payload:    debug.DefaultReceiverConfig.Payload,
				MaxHistory: pointer.Int(463),
			},
			valid: true,
		},

		{
			name: "maxhistory-negative",
			input: debug.ReceiverConfig{
				IntervalMs: debug.DefaultReceiverConfig.IntervalMs,
				Rounds:     debug.DefaultReceiverConfig.Rounds,
				Payload:    debug.DefaultReceiverConfig.Payload,
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
		config debug.ReceiverConfig
	}{
		{
			name:   "empty",
			config: debug.ReceiverConfig{},
		},

		{
			name:   "default",
			config: (&debug.ReceiverConfig{}).WithDefaults(),
		},

		{
			name:   "interval-zero",
			config: debug.ReceiverConfig{IntervalMs: pointer.Int(0)},
		},

		{
			name:   "interval-negative",
			config: debug.ReceiverConfig{IntervalMs: pointer.Int(-882)},
		},

		{
			name:   "interval-positive",
			config: debug.ReceiverConfig{IntervalMs: pointer.Int(1000)},
		},

		{
			name:   "rounds-zero",
			config: debug.ReceiverConfig{Rounds: pointer.Int(0)},
		},

		{
			name:   "rounds-negative",
			config: debug.ReceiverConfig{Rounds: pointer.Int(-234)},
		},

		{
			name:   "rounds-positive",
			config: debug.ReceiverConfig{Rounds: pointer.Int(2372)},
		},

		{
			name:   "payload-empty",
			config: debug.ReceiverConfig{Payload: pointer.String("")},
		},

		{
			name:   "payload-withvalue",
			config: debug.ReceiverConfig{Payload: pointer.String("some value")},
		},

		{
			name:   "maxhistory-zero",
			config: debug.ReceiverConfig{MaxHistory: pointer.Int(0)},
		},

		{
			name:   "maxhistory-positive",
			config: debug.ReceiverConfig{MaxHistory: pointer.Int(2347)},
		},

		{
			name:   "maxhistory-negative",
			config: debug.ReceiverConfig{MaxHistory: pointer.Int(-4)},
		},

		{
			name:   "maxhistory-maxint",
			config: debug.ReceiverConfig{MaxHistory: pointer.Int(9223372036854775807)},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			g := goldie.New(t, goldie.WithTestNameForDir(true))
			g.AssertJson(t, tc.name, tc.config)
		})
	}

}
