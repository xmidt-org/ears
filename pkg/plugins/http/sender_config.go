// SPDX-FileCopyrightText: 2025 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

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
