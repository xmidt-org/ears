// Copyright 2020 Comcast Cable Communications Management, LLC
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

package debug

import (
	"errors"
	"fmt"

	"github.com/xmidt-org/ears/pkg/config"
	"github.com/xmidt-org/ears/pkg/errs"
	"github.com/xmidt-org/ears/pkg/sender"
	"github.com/xmidt-org/ears/pkg/template"
	"github.com/xmidt-org/ears/pkg/validation"
	"github.com/xorcare/pointer"
)

func NewConfig(c interface{}) (*Config, error) {

	var cfg Config
	err := config.NewConfig(c, &cfg)

	if err != nil {
		return nil, &sender.InvalidConfigError{
			Err: err,
		}
	}

	return &cfg, nil
}

// WithDefaults will set
//   * The .Destination and .Writer if Destination is DestinationUnknown
//   * The .MaxHistory if it is nil
//
// The values will be set to the corresponding DefaultConfig.* values
//
func (sc *Config) WithDefaults() *Config {
	cfg := *sc

	if sc.Destination == DestinationUnknown {
		cfg.Destination = DefaultConfig.Destination
		cfg.Writer = DefaultConfig.Writer
	}

	if sc.MaxHistory == nil {
		cfg.MaxHistory = DefaultConfig.MaxHistory
	}

	return &cfg
}

// === Validation =========================================================

var _ validation.Validator = (*Config)(nil)

var configValidator *validation.Schema
var configSchema string

func (sc *Config) Schema() string {
	return configSchema
}

func (sc *Config) Validate() error {
	// JSON Schema is excellent at validating JSON.  This does not necessarily
	// translate well to Go structures that have values that cannot be instantiated
	// easily from a JSON string.  An example is an interface or a pointer to
	// i/o.
	//
	// Our `Writer` value is an interface and cannot be unmarshalled from a JSON
	// string.  Therefore, we will augment this validation by doing an additional
	// `Writer` value check.

	if configValidator == nil {
		return &validation.ProcessingError{
			Err: fmt.Errorf("validator is nil"),
		}
	}

	err := configValidator.Validate(sc)

	errs := []error{}

	var vErrs *validation.Errors
	var vErr *validation.Error
	if errors.As(err, &vErrs) {
		if vErrs.Errs != nil {
			errs = append(errs, vErrs.Errs...)
		}
	} else if errors.As(err, &vErr) {
		if vErr.Err != nil {
			errs = append(errs, vErr.Err)
		}
	}

	// Only one will be true
	switch {
	case sc.Destination == DestinationCustom && sc.Writer == nil:
		errs = append(errs, fmt.Errorf("writer cannot be nil if destination is 'custom'"))

	case sc.Destination != DestinationCustom && sc.Writer != nil:
		errs = append(errs, fmt.Errorf("if destination is not 'custom', writer should be nil"))
	}

	if len(errs) > 0 {
		return &validation.Errors{Errs: errs}
	}

	return nil
}

const senderSchemaTpl = `
{
  "$comment":"NOTE: Conditionals are only supported in draft-07",
  "$schema":"http://json-schema.org/draft-07/schema#",
  "$ref":"#/definitions/Config",
  "definitions":{
    "Config":{
      "title":"Config",
      "type":"object",
      "additionalProperties":false,
      "required":[
				"destination", "maxHistory"
      ],
      "properties":{
        "destination":{
          "enum": [
            "{{ .DestinationTypesStr | join "\", \"" }}"
          ],
          "default": "{{ .DestinationDefaultStr }}"
        },
        "maxHistory": {
          "type": "integer",
          "default": {{ .MaxHistoryDefault }},
          "minimum": {{ .MaxHistoryMinimum }}
        }
      }
    }
  }
}
`

func senderDestinationTypes() []string {
	types := []string{}

	for _, t := range DestinationTypeValues() {
		if t != DestinationUnknown {
			types = append(types, t.String())
		}
	}

	return types
}

// === Import / Export ==============================================

var _ config.Exporter = (*Config)(nil)
var _ config.Importer = (*Config)(nil)

func (c *Config) String() string {

	s, err := c.YAML()
	if err != nil {
		return errs.String("error", nil, err)
	}

	return s
}

func (c *Config) YAML() (string, error) {
	return config.ToYAML(c)
}

func (c *Config) FromYAML(in string) error {
	return config.FromYAML(in, c)
}

func (c *Config) JSON() (string, error) {
	return config.ToJSON(c)
}

func (c *Config) FromJSON(in string) error {
	return config.FromJSON(in, c)
}

func init() {
	// TODO:
	// * panic if our schema definition is bad?
	var err error

	configSchema, err = initConfigSchema()
	if err != nil {
		panic(&sender.InvalidConfigError{
			Err: err,
		})
	}

	err = initConfigValidator()
	if err != nil {
		panic(&sender.InvalidConfigError{
			Err: err,
		})
	}

	if configValidator == nil {
		panic(&sender.InvalidConfigError{
			Err: fmt.Errorf("Config validator still nil after intialization"),
		})
	}

}

func initConfigSchema() (string, error) {
	t, err := template.NewText()
	if err != nil {
		return "", err
	}

	err = t.Parse(senderSchemaTpl)
	if err != nil {
		return "", err
	}

	values := map[string]interface{}{
		"DestinationTypesStr":   senderDestinationTypes(),
		"DestinationDefault":    DefaultConfig.Destination,
		"DestinationDefaultStr": DefaultConfig.Destination.String(),
		"MaxHistoryDefault":     DefaultConfig.MaxHistory,
		"MaxHistoryMinimum":     pointer.Int(0),
		"WriterDefault":         nil,
	}

	result, err := t.Apply(values)
	if err != nil {
		return "", err
	}

	return result, nil
}

func initConfigValidator() error {
	// pull from the global variable to ensure
	// initialization worked correctly

	if configSchema == "" {
		return fmt.Errorf("Config schema is empty")
	}

	var err error
	configValidator, err = validation.NewSchema(configSchema)
	if err != nil {
		return fmt.Errorf("could not create Config validator: %w", err)
	}

	return nil
}
