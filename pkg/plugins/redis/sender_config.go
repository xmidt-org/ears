// SPDX-FileCopyrightText: 2025 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package redis

import (
	"fmt"

	"github.com/xeipuuv/gojsonschema"
)

// WithDefaults
func (sc SenderConfig) WithDefaults() SenderConfig {
	cfg := sc
	if cfg.Endpoint == "" {
		cfg.Endpoint = DefaultReceiverConfig.Endpoint
	}
	if cfg.Channel == "" {
		cfg.Channel = DefaultReceiverConfig.Channel
	}
	return cfg
}

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

const senderSchema = `
{
    "$schema": "http://json-schema.org/draft-06/schema#",
    "$ref": "#/definitions/SenderConfig",
    "definitions": {
        "SenderConfig": {
            "type": "object",
            "additionalProperties": false,
            "properties": {
                "endpoint": {
                    "type": "string"
                },
                "channel": {
                    "type": "string"
                }
            },
            "required": [
                "endpoint", "channel"
            ],
            "title": "SenderConfig"
        }
    }
}
`
