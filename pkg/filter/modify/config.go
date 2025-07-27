// SPDX-FileCopyrightText: 2025 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package modify

import (
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
	if c.Path == "" {
		cfg.Path = DefaultConfig.Path
	}
	if c.Paths == nil {
		cfg.Paths = DefaultConfig.Paths
	}
	if c.ToUpper == nil {
		cfg.ToUpper = DefaultConfig.ToUpper
	}
	if c.ToLower == nil {
		cfg.ToLower = DefaultConfig.ToLower
	}
	return &cfg
}

func (c *Config) Validate() error {
	return nil
}

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
