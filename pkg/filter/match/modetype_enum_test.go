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
	"fmt"
	"testing"

	"github.com/goccy/go-yaml"
	"github.com/xmidt-org/ears/pkg/filter/match"

	. "github.com/onsi/gomega"
)

// NOTE: Ideally we put in a pull request to go-enum
// to generate the test files as well to give 100% coverage
// for all genenerated code

func TestMatcherType(t *testing.T) {

	a := NewWithT(t)

	for _, e := range match.MatcherTypeValues() {
		m, err := match.ParseMatcherTypeString(e.String())
		a.Expect(err).To(BeNil())
		a.Expect(m).To(Equal(e))
		a.Expect(e.Registered()).To(BeTrue())

		a.Expect(
			match.MatcherTypeSliceContains(match.MatcherTypeValues(), e),
		).To(BeTrue())

		a.Expect(
			match.MatcherTypeSliceContainsAny(match.MatcherTypeValues(), e),
		).To(BeTrue())

		// Text
		{
			b, err := e.MarshalText()
			a.Expect(b).To(Equal([]byte(e.String())))
			a.Expect(err).To(BeNil())

			e2 := match.MatcherType(-1)
			err = (&e2).UnmarshalText(b)
			a.Expect(err).To(BeNil())
			a.Expect(e2).To(Equal(e))
		}

		// Binary
		{
			b, err := e.MarshalBinary()
			a.Expect(b).To(Equal([]byte(e.String())))
			a.Expect(err).To(BeNil())

			e2 := match.MatcherType(-1)
			err = (&e2).UnmarshalBinary(b)
			a.Expect(err).To(BeNil())
			a.Expect(e2).To(Equal(e))
		}

		// JSON
		{
			b, err := e.MarshalJSON()
			a.Expect(string(b)).To(Equal(fmt.Sprintf(`"%s"`, e.String())))
			a.Expect(err).To(BeNil())

			e2 := match.MatcherType(-1)
			err = (&e2).UnmarshalJSON(b)
			a.Expect(err).To(BeNil())
			a.Expect(e2).To(Equal(e))
		}

		// YAML
		{
			b, err := e.MarshalYAML()
			a.Expect(b).To(Equal(e.String()))
			a.Expect(err).To(BeNil())

			e2 := match.MatcherType(-1)
			err = yaml.Unmarshal([]byte(b.(string)), &e2)
			a.Expect(err).To(BeNil())
			a.Expect(e2).To(Equal(e))
		}

	}

}

func TestMatcherTypeErrors(t *testing.T) {

	a := NewWithT(t)

	// Invalid mode type value
	m := match.MatcherType(23)
	a.Expect(m.String()).To(Equal(`MatcherType(23)`))
	a.Expect(m.Registered()).To(BeFalse())
	a.Expect(
		match.MatcherTypeSliceContains(match.MatcherTypeValues(), m),
	).To(BeFalse())
	a.Expect(
		match.MatcherTypeSliceContainsAny(match.MatcherTypeValues(), m),
	).To(BeFalse())

	// Invalid mode type string
	m, err := match.ParseMatcherTypeString("invalid string goes here")
	a.Expect(m).To(Equal(match.MatcherType(0)))
	a.Expect(err).ToNot(BeNil())

	// Invalid JSON unmarshal
	err = m.UnmarshalJSON([]byte("}x4"))
	a.Expect(err).ToNot(BeNil())

	// Invalid YAML unmarshal
	err = m.UnmarshalYAML(func(interface{}) error { return fmt.Errorf("error") })
	a.Expect(err).ToNot(BeNil())

}
