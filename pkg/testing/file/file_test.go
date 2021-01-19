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

package file_test

import (
	"testing"

	"github.com/xmidt-org/ears/pkg/testing/file"

	. "github.com/onsi/gomega"
)

func TestGlob(t *testing.T) {

	a := NewWithT(t)

	files := file.Glob(t, "*.go")
	a.Expect(len(files)).To(Equal(2), "Expected 2 files to be returned")

}

func TestRead(t *testing.T) {
	testCases := []struct {
		file      string
		dir       string
		base      string
		extension string
		root      string
	}{
		{

			file:      "./file_test.go",
			dir:       "./",
			base:      "file_test.go",
			extension: "go",
			root:      "file_test",
		},
		{
			file:      "file.go",
			dir:       "",
			base:      "file.go",
			extension: "go",
			root:      "file",
		},

		// Have one of the letters in the extention match
		// some character on the left side of the `.`.
		// In this case, `l`.
		{
			file:      "/some/long/path/file.tpl.gol",
			dir:       "/some/long/path",
			base:      "file.tpl.gol",
			extension: "gol",
			root:      "file.tpl",
		},
	}

	for _, c := range testCases {
		t.Run(c.base, func(t *testing.T) {
			a := NewWithT(t)

			f := file.Read(t, c.file, file.OptionNotRequired)
			a.Expect(f.Base).To(Equal(c.base), "Base should be equal")
			a.Expect(f.Extension).To(Equal(c.extension), "Extension should be equal")
			a.Expect(f.Root).To(Equal(c.root), "Root should be equal")
		})
	}

}
