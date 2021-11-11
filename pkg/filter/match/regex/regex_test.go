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

package regex_test

import (
	"context"
	"testing"

	"github.com/xmidt-org/ears/pkg/event"

	"github.com/xmidt-org/ears/pkg/filter/match/regex"

	. "github.com/onsi/gomega"
)

func TestRegexPatternError(t *testing.T) {

	testCases := []struct {
		pattern string
	}{
		{pattern: `[\\`},
	}

	for _, tc := range testCases {
		t.Run(tc.pattern, func(t *testing.T) {
			a := NewWithT(t)
			m, err := regex.NewMatcher(tc.pattern, false)
			a.Expect(m).To(BeNil())
			a.Expect(err).ToNot(BeNil())
		})

	}

}

func TestRegexMatchError(t *testing.T) {
	ctx := context.Background()

	evt := func(payload interface{}) event.Event {
		e, _ := event.New(ctx, payload)
		return e
	}

	testCases := []struct {
		name  string
		event event.Event
	}{
		{
			name:  "<nil> event",
			event: nil,
		},
		{
			name:  "<nil> payload",
			event: evt(nil),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			a := NewWithT(t)
			m, err := regex.NewMatcher(".*", false)
			a.Expect(m).ToNot(BeNil())
			a.Expect(err).To(BeNil())

			a.Expect(m.Match(tc.event)).To(BeFalse())

		})

	}
}

func TestMatcherRegex(t *testing.T) {
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
	}

	for _, tc := range testCases {
		name := tc.name
		if name == "" {
			name = tc.pattern
		}

		ctx := context.Background()
		t.Run(name, func(t *testing.T) {
			a := NewWithT(t)

			m, err := regex.NewMatcher(tc.pattern, false)
			a.Expect(err).To(BeNil())

			for _, in := range tc.succeed {
				t.Run(in, func(t *testing.T) {
					a := NewWithT(t)
					e, err := event.New(ctx, in)
					a.Expect(err).To(BeNil())
					a.Expect(m.Match(e)).To(BeTrue(), "succeed input: "+in)
				})
			}

			for _, in := range tc.fail {
				t.Run(in, func(t *testing.T) {
					a := NewWithT(t)
					e, err := event.New(ctx, in)
					a.Expect(err).To(BeNil())
					a.Expect(m.Match(e)).To(BeFalse(), "fail input: "+in)
				})
			}

		})
	}

}
