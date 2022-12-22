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

package ws

import (
	"fmt"
	"github.com/xeipuuv/gojsonschema"
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
	if c.FromPath == "" {
		cfg.FromPath = DefaultConfig.FromPath
	}
	if c.ToPath == "" {
		cfg.ToPath = DefaultConfig.ToPath
	}
	if c.Method == "" {
		cfg.Method = DefaultConfig.Method
	}
	if c.Url == "" {
		cfg.Url = DefaultConfig.Url
	}
	if c.UrlPath == "" {
		cfg.UrlPath = DefaultConfig.UrlPath
	}
	if c.Body == "" {
		cfg.Body = DefaultConfig.Body
	}
	if c.Headers == nil {
		cfg.Headers = DefaultConfig.Headers
	}
	if c.EmptyPathValueRequired == nil {
		cfg.EmptyPathValueRequired = DefaultConfig.EmptyPathValueRequired
	}
	if c.Auth == nil {
		cfg.Auth = DefaultConfig.Auth
	}
	return &cfg
}

func (c *Config) Validate() error {
	schema := gojsonschema.NewStringLoader(filterSchema)
	doc := gojsonschema.NewGoLoader(*c)
	result, err := gojsonschema.Validate(schema, doc)
	if err != nil {
		return err
	}
	if !result.Valid() {
		return fmt.Errorf(fmt.Sprintf("%+v", result.Errors()))
	}
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

const filterSchema = `
{
    "$schema": "http://json-schema.org/draft-06/schema#",
    "$ref": "#/definitions/FilterConfig",
    "definitions": {
		"AuthConfig": {
            "type": "object",
            "additionalProperties": false,
            "properties": {
                "type": {
                    "type": "string",
					"enum": ["", "basic", "sat", "oauth", "oauth2"]
                },
                "username": {
                    "type": "string"
                },
                "password": {
                    "type": "string"
                },
                "clientId": {
                    "type": "string"
                },
                "clientSecret": {
                    "type": "string"
                },
                "tokenUrl": {
                    "type": "string"
                },
                "scopes": {
 					"type": "array",
      				"items": {
        				"type": "string"
      				}             
				}
            },
            "required": [
            ],
            "title": "AuthConfig"
		},
        "FilterConfig": {
            "type": "object",
            "additionalProperties": false,
            "properties": {
                "toPath": {
                    "type": "string"
                },
				"fromPath": {
                    "type": "string"
				},
				"url": {
                    "type": "string"
				},
				"urlPath": {
                    "type": "string"
				},
				"method": {
                    "type": "string",
					"enum": ["GET", "PUT", "POST", "DELETE"]
				},
				"body": {
                    "type": "string"
				},
				"headers": {
                    "type": "object"
				},
				"emptyPathValueRequired" : {
					"type": "boolean",
					"default": false
				},
				"auth": {
                    "$ref": "#/definitions/AuthConfig"
				}
            },
            "required": [
                "url",
				"toPath",
				"method"
            ],
            "title": "FilterConfig"
        }
    }
}
`
