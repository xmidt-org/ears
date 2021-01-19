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
