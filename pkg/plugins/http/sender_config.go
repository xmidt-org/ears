// Copyright 2021 Comcast Cable Communications Management, LLC
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

package http

import (
	"fmt"
	"github.com/xeipuuv/gojsonschema"
)

// Validate
func (sc *SenderConfig) Validate() error {
	schema := gojsonschema.NewStringLoader(senderSchema)
	doc := gojsonschema.NewGoLoader(*sc)
	result, err := gojsonschema.Validate(schema, doc)
	if err != nil {
		return err
	}
	if !result.Valid() {
		return fmt.Errorf(fmt.Sprintf("%+v", result.Errors()))
	}
	return nil
}

func (c SenderConfig) WithDefaults() SenderConfig {
	cfg := c
	if c.Method == "" {
		cfg.Method = DefaultSenderConfig.Method
	}
	if c.Url == "" {
		cfg.Url = DefaultSenderConfig.Url
	}
	if c.UrlPath == "" {
		cfg.UrlPath = DefaultSenderConfig.UrlPath
	}
	if c.Body == "" {
		cfg.Body = DefaultSenderConfig.Body
	}
	if c.Headers == nil {
		cfg.Headers = DefaultSenderConfig.Headers
	}
	if c.Auth == nil {
		cfg.Auth = DefaultSenderConfig.Auth
	}
	return cfg
}

const senderSchema = `
{
    "$schema": "http://json-schema.org/draft-06/schema#",
    "$ref": "#/definitions/SenderConfig",
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
                "grantType": {
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
        "SenderConfig": {
            "type": "object",
            "additionalProperties": false,
            "properties": {
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
				"auth": {
                    "$ref": "#/definitions/AuthConfig"
				}
            },
            "required": [
                "url",
				"method"
            ],
            "title": "SenderConfig"
        }
    }
}
`
