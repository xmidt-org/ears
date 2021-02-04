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

package template_test

import (
	"fmt"
	"testing"

	. "github.com/onsi/gomega"
	"github.com/xmidt-org/ears/pkg/template"
)

func TestWithOptions(t *testing.T) {
	testCases := []bool{false, true}

	for _, c := range testCases {
		t.Run(fmt.Sprint(c), func(t *testing.T) {
			a := NewWithT(t)

			p, err := template.NewText(
				template.WithIgnoreMissingValue(c),
			)

			a.Expect(err).To(BeNil())
			a.Expect(p).ToNot(BeNil())

			y, err := template.NewYAML(
				template.WithIgnoreMissingValue(c),
			)

			a.Expect(err).To(BeNil())
			a.Expect(y).ToNot(BeNil())

			j, err := template.NewJSON(
				template.WithIgnoreMissingValue(c),
			)

			a.Expect(err).To(BeNil())
			a.Expect(j).ToNot(BeNil())
		})

	}

}

func TestOptionToString(t *testing.T) {
	testCases := []struct {
		value    template.OptionMissingKey
		expected string
	}{
		{
			value:    template.OptionMissingKeyDefault,
			expected: "default",
		},
		{
			value:    template.OptionMissingKeyError,
			expected: "error",
		},
		{
			value:    template.OptionMissingKeyZero,
			expected: "zero",
		},
	}

	for _, c := range testCases {
		t.Run(fmt.Sprint(c), func(t *testing.T) {

			a := NewWithT(t)

			a.Expect(c.value.String()).To(Equal(c.expected))
			a.Expect(fmt.Sprint(c.value)).To(Equal(c.expected))
		})
	}
}
