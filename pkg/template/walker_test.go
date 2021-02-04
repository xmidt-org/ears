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

package template

import (
	"errors"
	"fmt"
	"path/filepath"
	"testing"

	. "github.com/onsi/gomega"
	"github.com/sebdah/goldie/v2"
)

func TestWalk(t *testing.T) {

	tpl := TestingNewTemplateEngine(OptionMissingKeyDefault)
	values := map[string]string{
		"foo":   "bar",
		"value": "oohhhh snap!",
	}

	sentence := "a sentence with a {{ .foo }} variable defined"

	serializer := func(tree interface{}) []byte {
		// We use fmt to serialize as (as of Go v1.12) it will
		// sort all keys to ease testing and examples
		// https://tip.golang.org/doc/go1.12#fmt
		//
		// Although it seemed like setting
		//		spew.Config.Sortkeys = true
		// would work, it didn't.
		return []byte(fmt.Sprintf("%#v", tree))
	}

	testCases := []struct {
		name       string
		tree       interface{}
		serializer func(interface{}) []byte
	}{
		{
			name: "string-withvar",
			tree: sentence,
		},
		{
			name: "string-pointer-withvar",
			tree: &sentence,
		},
		{
			name: "string-basic",
			tree: "a simple string",
		},
		{
			name: "bool-false",
			tree: false,
		},
		{
			name: "bool-true",
			tree: true,
		},
		{
			name: "int-zero",
			tree: 0,
		},
		{
			name: "int-positive",
			tree: 42,
		},
		{
			name: "array-string",
			tree: []string{"one", "two", "three"},
		},
		{
			name: "array-string-withvar",
			tree: []string{"one {{ .foo }}", "{{ .value }}", "three"},
		},
		{
			name: "array-string-withSprigFunctions",
			tree: []string{`{{ upper "foo" }}`, "{{ max 1 2 3 }}", `{{ sha256sum "Hello world!" }}`},
		},
		{
			name: "array-int",
			tree: []int{1, 2, 3, 4},
		},
		{
			name: "array-interface",
			tree: []interface{}{1, "two", true, 4.0},
		},
		{
			name: "array-interface-withvar",
			tree: []interface{}{1, "two {{ .foo }}", false, "{{ .somethingUndefined }}"},
		},
		{
			name: "map-string-stringWithvar",
			tree: map[string]string{
				"string":  "a",
				"withVar": "some {{ .value }} goes here",
			},
		},
		{
			name:       "map-interface-stringWithvar",
			serializer: serializer,
			tree: map[interface{}]string{
				"string":  "a",
				"withVar": "some {{ .value }} goes here",
			},
		},

		{
			name: "map-string-interfaceWithvar",
			tree: map[string]interface{}{
				"string":  "a",
				"int":     0,
				"float":   3.14,
				"withVar": "some {{ .value }} goes here",
			},
		},

		{
			name:       "map-interface-interfaceWithvar",
			serializer: serializer,
			tree: map[interface{}]interface{}{
				"string":  "a",
				"int":     0,
				"float":   3.14,
				"withVar": "some {{ .value }} goes here",
			},
		},

		{
			name: "map-stringvarkey-varValue",
			tree: map[string]interface{}{
				"{{ .foo }}":             "fooValue",
				"{{ .foo }}WithVarValue": "some {{ .value }} goes here",
			},
		},

		{
			name:       "map-interfaceVarkey-varValue",
			serializer: serializer,
			tree: map[interface{}]interface{}{
				"{{ .foo }}":             "fooValue",
				"{{ .foo }}WithVarValue": "some {{ .value }} goes here",
			},
		},
	}

	g := goldie.New(
		t,
		goldie.WithFixtureDir(filepath.Join(TestingDefaultFixtureDir, "golden")),
		goldie.WithTestNameForDir(true),
	)

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			a := NewWithT(t)
			err := walk(tpl, tc.tree, values)
			a.Expect(err).To(BeNil())

			if tc.serializer != nil {
				g.Assert(t, tc.name, tc.serializer(tc.tree))
			} else {
				g.AssertJson(t, tc.name, tc.tree)
			}
		})
	}

}

func TestWalkErrors(t *testing.T) {

	tpl := TestingNewTemplateEngine(OptionMissingKeyError)

	testCases := []struct {
		name string
		tree interface{}
		err  string
	}{
		{
			name: "array-parseError",
			tree: []string{"a bad {{ .value  template here "},
			err:  "ParseError",
		},
		{
			name: "array-missingKeyError",
			tree: []string{"no {{ .value }} to apply"},
			err:  "MissingKeyError",
		},
		{
			name: "array-missingFunction",
			tree: []string{"no {{ doit }} to apply"},
			err:  "ParseError",
		},

		// TODO:  Figure out how to trigger a ApplyValueError
	}
	g := goldie.New(
		t,
		goldie.WithFixtureDir(filepath.Join(TestingDefaultFixtureDir, "golden")),
		goldie.WithTestNameForDir(true),
	)

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			a := NewWithT(t)
			err := walk(tpl, tc.tree, nil)

			a.Expect(err).ToNot(BeNil())

			g.Assert(t, tc.name, []byte(err.Error()))

			switch tc.err {
			case "ParseError":
				var e *ParseError
				if !errors.As(err, &e) {
					t.Errorf(`Expected error of type %s, but got "%s"`, tc.err, err)
				}

			case "MissingKeyError":
				var e *MissingKeyError
				if !errors.As(err, &e) {
					t.Errorf(`Expected error of type %s, but got "%s"`, tc.err, err)
				}

			case "ApplyValueError":
				var e *ApplyValueError
				if !errors.As(err, &e) {
					t.Errorf(`Expected error of type %s, but got "%s"`, tc.err, err)
				}

			default:
				t.Errorf(`Unknown error of type %s specified`, tc.err)
			}
		})
	}

}
