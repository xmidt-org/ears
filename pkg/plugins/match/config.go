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
)

// WithDefaults will set
//   * The .Destination and .Writer if Destination is DestinationUnknown
//   * The .MaxHistory if it is nil
//
// The values will be set to the corresponding DefaultSenderConfig.* values
//
func (c *Config) WithDefaults() Config {
	if c == nil {
		return DefaultConfig
	}

	cfg := *c

	if c.Mode == ModeUnknown {
		cfg.Mode = DefaultConfig.Mode
	}

	if c.Matcher == MatcherUnknown {
		cfg.Matcher = DefaultConfig.Matcher
	}

	if c.Pattern == nil {
		cfg.Pattern = DefaultConfig.Pattern
	}

	return cfg
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
