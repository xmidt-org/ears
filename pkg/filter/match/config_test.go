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

package match_test

import (
	"testing"

	"github.com/xmidt-org/ears/pkg/filter/match"
	"github.com/xorcare/pointer"

	"github.com/sebdah/goldie/v2"

	. "github.com/onsi/gomega"
)

func TestConfigWithDefault(t *testing.T) {

	testCases := []struct {
		name     string
		input    match.Config
		expected match.Config
	}{
		{
			name:     "empty",
			input:    match.Config{},
			expected: match.DefaultConfig,
		},

		{
			name: "mode-allow",
			input: match.Config{
				Mode: match.ModeAllow,
			},
			expected: match.Config{
				Mode:            match.ModeAllow,
				Matcher:         match.DefaultConfig.Matcher,
				Pattern:         match.DefaultConfig.Pattern,
				ExactArrayMatch: pointer.Bool(true),
				Path:            "",
			},
		},

		{
			name: "mode-deny",
			input: match.Config{
				Mode: match.ModeDeny,
			},
			expected: match.Config{
				Mode:            match.ModeDeny,
				Matcher:         match.DefaultConfig.Matcher,
				Pattern:         match.DefaultConfig.Pattern,
				ExactArrayMatch: pointer.Bool(true),
				Path:            "",
			},
		},

		{
			name: "matcher-regex",
			input: match.Config{
				Matcher: match.MatcherRegex,
			},
			expected: match.Config{
				Mode:            match.DefaultConfig.Mode,
				Matcher:         match.MatcherRegex,
				Pattern:         match.DefaultConfig.Pattern,
				ExactArrayMatch: pointer.Bool(true),
				Path:            "",
			},
		},

		{
			name: "pattern",
			input: match.Config{
				Pattern: pointer.String("mypattern"),
			},
			expected: match.Config{
				Mode:            match.DefaultConfig.Mode,
				Matcher:         match.DefaultConfig.Matcher,
				Pattern:         pointer.String("mypattern"),
				ExactArrayMatch: pointer.Bool(true),
				Path:            "",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			a := NewWithT(t)

			d := tc.input.WithDefaults()
			a.Expect(*d).To(Equal(tc.expected))

		})
	}

}

func TestConfigValidation(t *testing.T) {
	testCases := []struct {
		name  string
		input match.Config
		valid bool
	}{
		{
			name:  "empty",
			input: match.Config{},
			valid: false,
		},

		{
			name:  "default",
			input: match.DefaultConfig,
			valid: true,
		},
		{
			name:  "with-defaults",
			input: *(match.Config{}.WithDefaults()),
			valid: true,
		},

		{
			name: "full",
			input: match.Config{
				Mode:    match.ModeAllow,
				Matcher: match.MatcherRegex,
				Pattern: pointer.String("fullpattern"),
			},
			valid: true,
		},

		{
			name: "mode-unset",
			input: match.Config{
				Mode:    match.ModeUnknown,
				Matcher: match.DefaultConfig.Matcher,
				Pattern: match.DefaultConfig.Pattern,
			},
			valid: false,
		},
		{
			name: "matcher-unset",
			input: match.Config{
				Mode:    match.DefaultConfig.Mode,
				Matcher: match.MatcherUnknown,
				Pattern: match.DefaultConfig.Pattern,
			},
			valid: false,
		},

		{
			name: "pattern-unset",
			input: match.Config{
				Mode:    match.DefaultConfig.Mode,
				Matcher: match.DefaultConfig.Matcher,
				Pattern: nil,
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

func TestConfigSerialization(t *testing.T) {
	testCases := []struct {
		name   string
		config match.Config
	}{
		{
			name:   "empty",
			config: match.Config{},
		},

		{
			name:   "default",
			config: *(match.Config{}.WithDefaults()),
		},

		{
			name:   "mode-allow",
			config: match.Config{Mode: match.ModeAllow},
		},

		{
			name:   "mode-deny",
			config: match.Config{Mode: match.ModeDeny},
		},

		{
			name:   "matcher-regex",
			config: match.Config{Matcher: match.MatcherRegex},
		},

		{
			name:   "pattern-empty",
			config: match.Config{Pattern: pointer.String("")},
		},

		{
			name:   "pattern-abcd1234",
			config: match.Config{Pattern: pointer.String("abcd1234")},
		},

		{
			name:   "pattern-strangecharacters",
			config: match.Config{Pattern: pointer.String(`💣🐢ab~!@#$%^&*)(_+-=}{;':"\|/?.,<>cd1234`)},
		},

		{
			name: "full",
			config: match.Config{
				Mode:    match.ModeAllow,
				Matcher: match.MatcherRegex,
				Pattern: pointer.String(`^[[:alnum:]]{4,12}`),
				Path:    "",
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
