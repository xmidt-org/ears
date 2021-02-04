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
	"path/filepath"
	"strings"
	"testing"

	"github.com/sebdah/goldie/v2"
	"github.com/xmidt-org/ears/pkg/testing/file"

	. "github.com/onsi/gomega"
	"github.com/xmidt-org/ears/pkg/template"
)

func TestJSONParseFile(t *testing.T) {
	g := goldie.New(
		t,
		goldie.WithFixtureDir(filepath.Join(template.TestingDefaultFixtureDir, "golden")),
		goldie.WithTestNameForDir(true),
	)

	files := file.Glob(t, filepath.Join(template.TestingDefaultFixtureDir, "*.json"))

	for _, f := range files {
		t.Run(f.Base, func(t *testing.T) {
			a := NewWithT(t)

			expectError := strings.Contains(f.Base, "fail")

			p, err := template.NewJSON(
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

func TestJSONApply(t *testing.T) {
	files := file.Glob(t, filepath.Join(template.TestingDefaultFixtureDir, "pass-*.json"))

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

			p, err := template.NewJSON(
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
