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
	"fmt"

	"github.com/xmidt-org/ears/pkg/config"
	"github.com/xmidt-org/ears/pkg/errs"
	"github.com/xmidt-org/ears/pkg/receiver"
	"github.com/xmidt-org/ears/pkg/validation"
)

func NewConfig(c interface{}) (*Config, error) {

	var cfg Config
	err := config.NewConfig(c, &cfg)

	if err != nil {
		return nil, &receiver.InvalidConfigError{
			Err: err,
		}
	}

	return &cfg, nil
}

// WithDefaults returns a new config object that has all
// of the unset (nil) values filled in.
func (rc *Config) WithDefaults() *Config {
	cfg := *rc

	if cfg.IntervalMs == nil {
		cfg.IntervalMs = DefaultConfig.IntervalMs
	}

	if cfg.Rounds == nil {
		cfg.Rounds = DefaultConfig.Rounds
	}

	if cfg.Payload == nil {
		cfg.Payload = DefaultConfig.Payload
	}

	if cfg.MaxHistory == nil {
		cfg.MaxHistory = DefaultConfig.MaxHistory
	}

	return &cfg
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

// === Validation =====================================================

var _ validation.Validator = (*Config)(nil)

func (rc *Config) Schema() string {
	return configSchema
}

// Validate returns an error upon validation failure
func (rc *Config) Validate() error {

	if configValidator == nil {
		return &validation.ProcessingError{
			Err: fmt.Errorf("validator is nil"),
		}
	}

	return configValidator.Validate(rc)
}

var configValidator *validation.Schema

const configSchema = `
{
    "$schema": "http://json-schema.org/draft-07/schema#",
    "$ref": "#/definitions/Config",
    "definitions": {
        "Config": {
            "type": "object",
            "additionalProperties": false,
            "properties": {
                "intervalMs": {
                    "type": "integer",
										"minimum": 1
                },
                "rounds": {
                    "type": "integer",
										"minimum": -1
                },
                "payload": {
                    "type": ["string", "object"]
                },
                "maxHistory": {
                    "type": "integer",
										"minimum": 0
                }
            },
            "required": [
                "intervalMs",
                "maxHistory",
                "rounds"
            ],
            "title": "Config"
        }
    }
}
`

// === Initialization =================================================

func init() {
	// TODO:
	// * panic if our schema definition is bad?

	var err error
	err = initConfigValidator()
	if err != nil {
		panic(&receiver.InvalidConfigError{
			Err: err,
		})
	}

}

func initConfigValidator() error {
	// pull from the global variable to ensure
	// initialization worked correctly

	var err error
	configValidator, err = validation.NewSchema(configSchema)
	if err != nil {
		return fmt.Errorf("could not create Config validator: %w", err)
	}

	return nil
}
