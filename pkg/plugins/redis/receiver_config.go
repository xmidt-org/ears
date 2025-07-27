// SPDX-FileCopyrightText: 2025 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package redis

import (
	"fmt"

	"github.com/xeipuuv/gojsonschema"
)

// WithDefaults returns a new config object that has all
// of the unset (nil) values filled in.
func (rc *ReceiverConfig) WithDefaults() ReceiverConfig {
	cfg := *rc
	if cfg.Endpoint == "" {
		cfg.Endpoint = DefaultReceiverConfig.Endpoint
	}
	if cfg.Channel == "" {
		cfg.Channel = DefaultReceiverConfig.Channel
	}
	if cfg.TracePayloadOnNack == nil {
		cfg.TracePayloadOnNack = DefaultReceiverConfig.TracePayloadOnNack
	}
	return cfg
}

// Validate returns an error upon validation failure
func (rc *ReceiverConfig) Validate() error {
	schema := gojsonschema.NewStringLoader(receiverSchema)
	doc := gojsonschema.NewGoLoader(*rc)
	result, err := gojsonschema.Validate(schema, doc)
	if err != nil {
		return err
	}
	if !result.Valid() {
		return fmt.Errorf(fmt.Sprintf("%+v", result.Errors()))
	}
	return nil
}

const receiverSchema = `
{
    "$schema": "http://json-schema.org/draft-06/schema#",
    "$ref": "#/definitions/ReceiverConfig",
    "definitions": {
        "ReceiverConfig": {
            "type": "object",
            "additionalProperties": false,
            "properties": {
                "endpoint": {
                    "type": "string"
                },
                "channel": {
                    "type": "string"
                },
				"tracePayloadOnNack" : {
					"type": "boolean",
					"default": false
				}
            },
            "required": [
                "endpoint", "channel"
            ],
            "title": "ReceiverConfig"
        }
    }
}
`
