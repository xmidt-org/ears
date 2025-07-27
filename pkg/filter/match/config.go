// SPDX-FileCopyrightText: 2025 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

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

func (c Config) WithDefaults() *Config {
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
	if c.ExactArrayMatch == nil {
		cfg.ExactArrayMatch = DefaultConfig.ExactArrayMatch
	}
	return &cfg
}

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
