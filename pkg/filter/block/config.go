// SPDX-FileCopyrightText: 2025 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package block

import (
	"github.com/xmidt-org/ears/pkg/config"
	"github.com/xmidt-org/ears/pkg/errs"
	"github.com/xmidt-org/ears/pkg/validation"
)

var _ config.Exporter = (*Config)(nil)
var _ config.Importer = (*Config)(nil)
var _ validation.Validator = (*Config)(nil)

func NewConfig(config interface{}) (*Config, error) {
	return &Config{}, nil
}

type Config struct{}

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

func (c *Config) Validate() error {
	return nil
}
