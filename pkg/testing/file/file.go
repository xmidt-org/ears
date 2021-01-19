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

// Package file assists in loading test files from the file system.
// Errors will call t.Error(err).  Passing in `OptionNotRequired`
// will cause the function to fail if no file is found.
//
package file

import (
	"io/ioutil"
	"path/filepath"
	"strings"
	"testing"
)

type File struct {
	Path string
	Dir  string
	Base string

	// Extension will NOT include the preceding "."
	Extension string
	Root      string
	Data      string
}

type Option int

const (
	OptionUndefined Option = iota
	OptionNotRequired
)

// Glob takes in a glob pattern and will return the path, base file
// name, and read in contents.  Will fail the test if the path
// pattern is bad.  If OptionNotRequired is passed in, we will
// not fail if no files were found.
func Glob(t *testing.T, pattern string, options ...Option) []File {
	var files []File

	matches, err := filepath.Glob(pattern)
	if err != nil {
		t.Error(err)
	}

	for _, f := range matches {
		files = append(
			files,
			Read(t, f),
		)
	}

	if len(files) == 0 && !listHasOption(options, OptionNotRequired) {
		t.Errorf("No files found matching %s", pattern)
	}

	return files
}

// Read takes in a file path and returns a single File object.  It will
// fail the test if the file is not found, unless OptionNotRequired
// is passed in.
func Read(t *testing.T, path string, options ...Option) File {

	data, err := ioutil.ReadFile(path)
	if err != nil && !listHasOption(options, OptionNotRequired) {
		t.Error(err)
	}

	base := filepath.Base(path)
	ext := strings.TrimPrefix(filepath.Ext(path), ".")
	root := strings.TrimSuffix(base, "."+ext)

	return File{
		Path:      path,
		Dir:       filepath.Dir(path),
		Base:      base,
		Extension: ext,
		Root:      root,
		Data:      string(data),
	}

}

func listHasOption(options []Option, item Option) bool {
	for _, o := range options {
		if o == item {
			return true
		}
	}

	return false
}
