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

package _example

import (
	"github.com/xmidt-org/ears/pkg/config"
	"github.com/xmidt-org/ears/pkg/errs"
	"github.com/xmidt-org/ears/pkg/filter"
	"github.com/xmidt-org/ears/pkg/validation"
)

// NewConfig will return the custom Config object.  The helper
// `pkgconfig.NewConfig` does the hard work of converting
// the input from:
//
//   * yaml
//   * json
//   * Config{}
//   * &Config{}
//
// into a *Config{} object.
func NewConfig(c interface{}) (*Config, error) {

	var cfg Config
	err := config.NewConfig(c, &cfg)

	if err != nil {
		return nil, &filter.InvalidConfigError{
			Err: err,
		}
	}

	return &cfg, nil
}

var _ config.Exporter = (*Config)(nil)
var _ config.Importer = (*Config)(nil)
var _ validation.Validator = (*Config)(nil)

// Config is the structure that defines your configuration.
// It should specify JSON tags for easy serialization
// & deserialization.
type Config struct{}

// WithDefaults constructs a new Config object and will
// set any values for any values that are missing.  A
// common practice is to have the member variables of a
// struct to always be pointers (e.g: *int) so that
// they can be `nil`, which indicates "not supplied."
// It would be these `nil` values that will get filled in.
func (c Config) WithDefaults() *Config {
	cfg := c
	return &cfg
}

// Validate will return validation errors if the configuration
// has invalid values.  NOTE: Make sure to run `WithDefaults()`
// prior to validation as `nil` (not supplied) values usually
// are not valid configuration values.
func (c *Config) Validate() error {
	return nil
}

// =============================================================
// Import / Export Helper Functions
// =============================================================

// String turns the config into a readable string
func (c *Config) String() string {
	s, err := c.YAML()
	if err != nil {
		return errs.String("error", nil, err)
	}

	return s
}

// YAML exports the config as YAML.
func (c *Config) YAML() (string, error) {
	return config.ToYAML(c)
}

// FromYAML will consume YAML and will set the
// values of the structure to the values provided
// in the input string.
func (c *Config) FromYAML(in string) error {
	return config.FromYAML(in, c)
}

// JSON exports the config as JSON
func (c *Config) JSON() (string, error) {
	return config.ToJSON(c)
}

// FromJSON will consume JSON and will set the
// values of the structure to the values provided
// in the input string.
func (c *Config) FromJSON(in string) error {
	return config.FromJSON(in, c)
}
