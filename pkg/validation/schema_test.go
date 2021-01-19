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

package validation_test

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/goccy/go-yaml"
	. "github.com/onsi/gomega"
	"github.com/xmidt-org/ears/pkg/testing/file"
	"github.com/xmidt-org/ears/pkg/validation"
)

const testfiledir = "testdata"

func TestSchemaValidator(t *testing.T) {

	schemas := file.Glob(t, filepath.Join(testfiledir, `*.schema.yaml`))

	for _, s := range schemas {
		parts := strings.Split(s.Base, ".")
		baseName := parts[0]

		t.Run(baseName, func(t *testing.T) {

			a := NewWithT(t)
			v, err := validation.NewSchema(s.Data)
			a.Expect(err).To(BeNil())
			a.Expect(v.Schema()).To(Equal(s.Data))

			successFiles := file.Glob(t,
				filepath.Join(testfiledir, baseName+`.success.*.yaml`),
				file.OptionNotRequired,
			)

			failFiles := file.Glob(t,
				filepath.Join(testfiledir, baseName+`.fail.*.yaml`),
				file.OptionNotRequired,
			)

			testCases := []struct {
				name        string
				files       []file.File
				expectError bool
			}{
				{
					name:        "success",
					files:       successFiles,
					expectError: false,
				},
				{
					name:        "fail",
					files:       failFiles,
					expectError: true,
				},
			}

			e := func(a *WithT, err error, expectError bool, msg string) {
				if expectError {
					a.Expect(err).ToNot(BeNil(), msg)
				} else {
					a.Expect(err).To(BeNil(), msg)
				}
			}

			for _, tc := range testCases {

				t.Run(tc.name, func(t *testing.T) {

					for _, f := range tc.files {
						parts := strings.Split(f.Base, ".")
						testName := parts[2]

						t.Run(testName, func(t *testing.T) {
							a := NewWithT(t)

							err := v.Validate(f.Data)
							e(a, err, tc.expectError, f.Base)

							err = v.Validate([]byte(f.Data))
							e(a, err, tc.expectError, f.Base)

							d := map[string]interface{}{}
							err = yaml.Unmarshal([]byte(f.Data), &d)
							a.Expect(err).To(BeNil())
							err = v.Validate(d)
							e(a, err, tc.expectError, f.Base)
						})
					}

				})
			}

		})

	}

}
