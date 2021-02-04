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
	"path/filepath"
	"strings"
	"testing"

	"github.com/sebdah/goldie/v2"
	"github.com/xmidt-org/ears/pkg/testing/file"

	. "github.com/onsi/gomega"
	"github.com/xmidt-org/ears/pkg/template"
)

func TestTextParseFile(t *testing.T) {
	g := goldie.New(
		t,
		goldie.WithFixtureDir(filepath.Join(template.TestingDefaultFixtureDir, "golden")),
		goldie.WithTestNameForDir(true),
	)

	files := file.Glob(t, filepath.Join(template.TestingDefaultFixtureDir, "*.txt"))

	for _, f := range files {
		t.Run(f.Base, func(t *testing.T) {
			a := NewWithT(t)

			expectError := strings.Contains(f.Base, "fail")

			p, err := template.NewText(
				template.WithIgnoreMissingValue(true),
			)

			a.Expect(err).To(BeNil())
			err = p.Parse(f.Data)
			if expectError {
				a.Expect(err).NotTo(BeNil())
				g.Assert(t, f.Base, []byte(err.Error()))
			} else {
				a.Expect(err).To(BeNil())
			}
		})
	}

}

func TestTextMissingKeyOptions(t *testing.T) {
	files := file.Glob(t, filepath.Join(template.TestingDefaultFixtureDir, "pass-withvars*.txt"))

	g := goldie.New(t,
		goldie.WithFixtureDir(filepath.Join(template.TestingDefaultFixtureDir, "golden")),
		goldie.WithTestNameForDir(true),
	)

	for _, option := range []template.OptionMissingKey{
		template.OptionMissingKeyDefault,
		template.OptionMissingKeyZero,
		template.OptionMissingKeyError,
	} {
		for _, f := range files {
			testName := f.Base + "-" + option.String()
			t.Run(testName, func(t *testing.T) {
				a := NewWithT(t)

				p, err := template.NewText(
					template.WithMissingKeyAction(option),
				)

				a.Expect(err).To(BeNil())
				err = p.Parse(f.Data)
				a.Expect(err).To(BeNil())

				content, err := p.Apply(nil)
				if option != template.OptionMissingKeyError {
					a.Expect(err).To(BeNil())
				} else {
					a.Expect(err).ToNot(BeNil())
					content = err.Error()
				}

				g.Assert(t, testName, []byte(content))

			})
		}
	}
}

func TestTextWithIgnoreMissingValue(t *testing.T) {
	files := file.Glob(t, filepath.Join(template.TestingDefaultFixtureDir, "pass-withvars*.txt"))

	g := goldie.New(t,
		goldie.WithFixtureDir(filepath.Join(template.TestingDefaultFixtureDir, "golden")),
		goldie.WithTestNameForDir(true),
	)

	for _, option := range []bool{false, true} {
		for _, f := range files {
			testName := f.Base + "-" + fmt.Sprint(option)
			t.Run(testName, func(t *testing.T) {
				a := NewWithT(t)

				p, err := template.NewText(
					template.WithIgnoreMissingValue(option),
				)

				a.Expect(err).To(BeNil())
				err = p.Parse(f.Data)
				a.Expect(err).To(BeNil())

				content, err := p.Apply(nil)
				if option == true {
					a.Expect(err).To(BeNil())
				} else {
					a.Expect(err).ToNot(BeNil())
					content = err.Error()
				}

				g.Assert(t, testName, []byte(content))

			})
		}
	}
}

func TestTextApply(t *testing.T) {
	files := file.Glob(t, filepath.Join(template.TestingDefaultFixtureDir, "pass-*.txt"))

	values := map[string]string{
		"someVarKey": "replacementVarKey",
		"Action":     "oohhhh snap!",
		"Foo":        "bars",
		"value":      "replacement value here",
	}

	g := goldie.New(
		t,
		goldie.WithFixtureDir(filepath.Join(template.TestingDefaultFixtureDir, "golden")),
		goldie.WithTestNameForDir(true),
	)

	for _, f := range files {
		t.Run(f.Base, func(t *testing.T) {
			a := NewWithT(t)

			expectError := strings.HasSuffix(f.Root, "evalerr")

			p, err := template.NewText(
				template.WithIgnoreMissingValue(true),
			)
			a.Expect(err).To(BeNil())

			err = p.Parse(f.Data)
			a.Expect(err).To(BeNil())

			result, err := p.Apply(values)

			if expectError {
				a.Expect(err).NotTo(BeNil())
				result = err.Error()
			} else {
				a.Expect(err).To(BeNil())
			}

			g.Assert(t, f.Base, []byte(result))

		})
	}

}
