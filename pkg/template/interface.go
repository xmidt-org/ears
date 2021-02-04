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

// ==============================================================
// Errors that may be returned
// ==============================================================

// templateError interface makes sure each error
// supports the Go v1.13 error interface
type templateError interface {
	error
	Unwrap() error
}

// ParseError is returned when there was an issue parsing the
// given content (usually some bad template syntax)
type ParseError struct {
	templateError
	Err     error
	Content string
}

// MissingKeyError is returned when OptionMissingKeyError is
// chosen for the parser and a value for the template is missing.
// Values is all the values that were passed in.
type MissingKeyError struct {
	templateError
	Err     error
	Key     string
	Content string
	Values  interface{}
}

// ApplyValueError occurrs when
type ApplyValueError struct {
	templateError
	Err     error
	Content string
	Values  interface{}
}

// ==============================================================
// Template Processor
// ==============================================================

type Processor interface {
	Parse(data string) error
	Apply(values interface{}) (string, error)
}

// New returns a text template processor.  This will allow you
// to take any simple template contents and apply values to it
func NewText(options ...Options) (*text, error) {
	p := text{}

	for _, option := range options {
		option(&p)
	}

	p.init()

	return &p, nil
}

// NewYAML returns a processor that is able to properly maintain
// YAML formatting.  This has been an issue in the past with trying
// to inject certificates into a document that has been interpreted
// as text.  This processor will parse the YAML document, will
// substitute values, and then will output the new YAML document.
// NOTE:  Key order will change from the original document.
func NewYAML(options ...Options) (*yaml, error) {
	p := yaml{}

	for _, option := range options {
		option(&p)
	}

	p.init()

	return &p, nil
}

// NewJSON returns a processor that is able to properly maintain
// JSON formatting.  This processor will parse the JSON document, will
// substitute values, and then will output the new JSON document.
// NOTE:  Key order will change from the original document.
func NewJSON(options ...Options) (*json, error) {
	p := json{}

	for _, option := range options {
		option(&p)
	}

	p.init()

	return &p, nil
}

// ==============================================================
// Processor Options
// ==============================================================

type Options func(options) error

type OptionMissingKey int

const (
	OptionMissingKeyDefault OptionMissingKey = iota
	OptionMissingKeyError
	OptionMissingKeyZero
)

func (o OptionMissingKey) String() string {
	switch o {
	case OptionMissingKeyError:
		return "error"
	case OptionMissingKeyZero:
		return "zero"
	default:
		return "default"
	}
}

// WithIgnoreMissingValue will ignore any template processing errors
// if `true` is passed in.
func WithIgnoreMissingValue(shouldIgnore bool) Options {
	return func(o options) error {
		mka := OptionMissingKeyDefault
		if !shouldIgnore {
			mka = OptionMissingKeyError
		}

		return o.withMissingKeyAction(mka)
	}
}

func WithMissingKeyAction(action OptionMissingKey) Options {
	return func(o options) error {
		return o.withMissingKeyAction(action)
	}
}
