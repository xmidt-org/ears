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

package match_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/goccy/go-yaml"
	"github.com/xmidt-org/ears/pkg/event"
	"github.com/xorcare/pointer"

	"github.com/xmidt-org/ears/pkg/filter/match"

	plugmatch "github.com/xmidt-org/ears/pkg/plugins/match"

	. "github.com/onsi/gomega"
)

func TestFilterRegex(t *testing.T) {
	testCases := []struct {
		name    string
		pattern string
		succeed []string
		fail    []string
	}{

		// Original tests from regex filter
		{
			pattern: `^.*$`,
			succeed: []string{""},
		},

		{
			pattern: `^.+$`,
			succeed: []string{"a"},
			fail:    []string{""},
		},

		{
			pattern: `^hello(world|universe)$`,
			succeed: []string{"helloworld"},
			fail:    []string{"goodbyeworld"},
		},

		{
			pattern: `^[[:alnum:]]+$`,
			succeed: []string{"abcDEFG1234567890"},
			fail:    []string{"abc_&@_12345"},
		},

		// Add additional test cases
		//  * https://golang.org/pkg/regexp/syntax/
		//  * https://regexr.com/

		{
			pattern: `^a{3,5}$`,
			succeed: []string{"aaa", "aaaa", "aaaaa"},
			fail:    []string{"", "a", "aaaaaa"},
		},

		{
			pattern: `[\d]`,
			succeed: []string{"1", "23", "657567", "abcd123324sdfsd"},
			fail:    []string{"", "sdf", "@#$%"},
		},

		{
			pattern: `[\d]{4}`,
			succeed: []string{"1324", "1233sdfsd", "sdfj 23 sdfkjsdf 2344 sdfs"},
			fail:    []string{"", "sdf", "1 12 123 49"},
		},

		{
			pattern: `^[\d]{4}$`,
			succeed: []string{"1324"},
			fail:    []string{"", " 1234", "1234 ", "12345"},
		},
	}

	a := NewWithT(t)
	p, err := plugmatch.NewPlugin()
	a.Expect(err).To(BeNil())

	for _, tc := range testCases {
		name := tc.name
		if name == "" {
			name = tc.pattern
		}

		ctx := context.Background()
		t.Run(name, func(t *testing.T) {
			a := NewWithT(t)

			for _, mode := range []match.ModeType{match.ModeAllow, match.ModeDeny} {

				f, err := p.NewFilterer(match.Config{
					Mode:    mode,
					Matcher: match.MatcherRegex,
					Pattern: &tc.pattern,
				})
				a.Expect(err).To(BeNil())

				succeed := tc.succeed
				fail := tc.fail

				if mode == match.ModeDeny {
					// We flip the values as we expect the opposite
					succeed = tc.fail
					fail = tc.succeed
				}

				for _, in := range succeed {
					t.Run(mode.String()+"/"+in, func(t *testing.T) {
						a := NewWithT(t)
						e, err := event.NewEvent(ctx, in)
						a.Expect(err).To(BeNil())

						evts, err := f.Filter(e)
						a.Expect(err).To(BeNil())
						a.Expect(evts).To(HaveLen(1), fmt.Sprintf(
							"mode: %s, succeed input: %s", mode.String(), in,
						))
					})

				}

				for _, in := range fail {
					t.Run(mode.String()+"/"+in, func(t *testing.T) {
						a := NewWithT(t)

						e, err := event.NewEvent(ctx, in)
						a.Expect(err).To(BeNil())

						evts, err := f.Filter(e)
						a.Expect(err).To(BeNil())
						a.Expect(evts).To(HaveLen(0), fmt.Sprintf(
							"mode: %s, fail input: %s", mode.String(), in,
						))

					})
				}
			}

		})
	}

}

func TestNewFilterConfig(t *testing.T) {
	testCases := []struct {
		name     string
		config   match.Config
		expected match.Config
	}{

		// empty is not a good test.  When it comes in as a blank string, it
		// will end up having its values set to defaults
		{
			name:     "empty",
			config:   match.Config{},
			expected: match.DefaultConfig,
		},

		{
			name:     "default",
			config:   match.DefaultConfig,
			expected: match.DefaultConfig,
		},

		{
			name:     "with-defaults",
			config:   *(match.Config{}.WithDefaults()),
			expected: *(match.Config{}.WithDefaults()),
		},
	}

	a := NewWithT(t)
	p, err := plugmatch.NewPlugin()
	a.Expect(err).To(BeNil())

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			a := NewWithT(t)

			y, err := yaml.Marshal(tc.config)
			a.Expect(err).To(BeNil())
			yStr := string(y)

			exp, err := yaml.Marshal(tc.expected)
			a.Expect(err).To(BeNil())
			expectedStr := string(exp)

			configCases := []struct {
				name   string
				config interface{}
			}{
				{
					name:   "string",
					config: yStr,
				},

				{
					name:   "bytes",
					config: y,
				},

				{
					name:   "value",
					config: tc.config,
				},

				{
					name:   "pointer",
					config: &tc.config,
				},

				{
					name:   "invalid",
					config: 987,
				},
			}

			for _, cc := range configCases {
				t.Run(cc.name, func(t *testing.T) {

					a := NewWithT(t)

					f, err := p.NewFilterer(cc.config)
					if cc.name == "invalid" {
						a.Expect(err).ToNot(BeNil())
						a.Expect(f).To(BeNil())
						return
					}

					a.Expect(err).To(BeNil())
					a.Expect(f).ToNot(BeNil())

					fcfg, ok := f.(interface {
						Config() match.Config
					})
					a.Expect(ok).To(BeTrue(), "filter supports Config interface")
					a.Expect(fcfg).ToNot(BeNil())

					cfg := fcfg.Config()

					a.Expect(cfg.String()).To(Equal(expectedStr))

				})

			}

		})

	}

}

func TestNewFilterBadConfig(t *testing.T) {

	a := NewWithT(t)
	p, err := plugmatch.NewPlugin()
	a.Expect(err).To(BeNil())

	{
		f, err := p.NewFilterer(match.Config{
			Mode:    match.ModeType(93),
			Matcher: match.MatcherRegex,
			Pattern: pointer.String("pattern"),
		})
		a.Expect(err).ToNot(BeNil())
		a.Expect(f).To(BeNil())
	}

	{
		f, err := p.NewFilterer(match.Config{
			Mode:    match.ModeAllow,
			Matcher: match.MatcherType(34),
			Pattern: pointer.String("pattern"),
		})
		a.Expect(err).ToNot(BeNil())
		a.Expect(f).To(BeNil())
	}

}

func TestNilFilter(t *testing.T) {
	var f *match.Filter

	a := NewWithT(t)

	_, err := f.Filter(nil)
	a.Expect(err).ToNot(BeNil())

	a.Expect(f.Config()).To(Equal(match.Config{}))
}
