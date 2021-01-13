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

package match

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/xmidt-org/ears/pkg/config"
	pkgconfig "github.com/xmidt-org/ears/pkg/config"
	"github.com/xmidt-org/ears/pkg/errs"
	"github.com/xmidt-org/ears/pkg/filter"
)

func NewConfig(config interface{}) (*Config, error) {

	var cfg Config
	err := pkgconfig.NewConfig(config, &cfg)

	if err != nil {
		return nil, &filter.InvalidConfigError{
			Err: err,
		}
	}

	return &cfg, nil
}

// WithDefaults will set
//   * The .Destination and .Writer if Destination is DestinationUnknown
//   * The .MaxHistory if it is nil
//
// The values will be set to the corresponding DefaultSenderConfig.* values
//

func (c Config) WithDefaults() *Config {
	// if c == nil {
	// 	return &DefaultConfig
	// }

	cfg := c

	if c.Mode == ModeUnknown {
		cfg.Mode = DefaultConfig.Mode
	}

	if c.Matcher == MatcherUnknown {
		cfg.Matcher = DefaultConfig.Matcher
	}

	if c.Pattern == nil {
		cfg.Pattern = DefaultConfig.Pattern
	}

	return &cfg
}

// Validate will ensure that:
//   * The .Destination is a valid DestinationType and is not DestinationUnknown
//   * Max history is >= 0
//   * The .Writer is only set when .Destination is DestinationCustom
func (c *Config) Validate() error {
	s := *c

	// Allow this list to easily expand over time
	validModes := []interface{}{}
	for _, t := range ModeTypeValues() {
		if t != ModeUnknown {
			validModes = append(validModes, t)
		}
	}

	validMatchers := []interface{}{}
	for _, t := range MatcherTypeValues() {
		if t != MatcherUnknown {
			validMatchers = append(validMatchers, t)
		}
	}

	return validation.ValidateStruct(&s,
		validation.Field(&s.Mode,
			validation.Required,
			validation.In(validModes...),
		),

		validation.Field(&s.Matcher,
			validation.Required,
			validation.In(validMatchers...),
		),

		validation.Field(&s.Pattern,
			validation.NotNil,
		),
	)

}

// Exporter interface

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
