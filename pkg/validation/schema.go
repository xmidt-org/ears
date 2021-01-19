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

package validation

import (
	"fmt"

	"github.com/xeipuuv/gojsonschema"
)

type Schema struct {
	schema   string
	compiled gojsonschema.JSONLoader
}

func NewSchema(schema string) (*Schema, error) {
	return &Schema{
		schema:   schema,
		compiled: gojsonschema.NewStringLoader(schema),
	}, nil
}

func (s *Schema) Schema() string {
	return s.schema
}

func (s *Schema) Validate(data interface{}) error {
	var doc gojsonschema.JSONLoader

	switch d := data.(type) {
	case string:
		doc = gojsonschema.NewStringLoader(d)
	case []byte:
		doc = gojsonschema.NewStringLoader(string(d))
	default:
		doc = gojsonschema.NewGoLoader(d)
	}

	result, err := gojsonschema.Validate(s.compiled, doc)

	if err != nil {
		return &ProcessingError{Err: err}
	}

	if !result.Valid() {
		errs := []error{}
		for _, e := range result.Errors() {
			errs = append(errs, fmt.Errorf(e.String()))
		}
		return &Errors{
			Errs: errs,
		}
	}

	return nil

}
