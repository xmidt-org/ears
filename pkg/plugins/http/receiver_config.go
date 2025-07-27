// SPDX-FileCopyrightText: 2025 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package http

import (
	"fmt"

	"github.com/xeipuuv/gojsonschema"
)

// WithDefaults
func (rc *ReceiverConfig) WithDefaults() ReceiverConfig {
	cfg := *rc
	if cfg.SuccessStatus == nil {
		cfg.SuccessStatus = DefaultReceiverConfig.SuccessStatus
	}
	if cfg.FailureStatus == nil {
		cfg.FailureStatus = DefaultReceiverConfig.FailureStatus
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
                "path": {
                    "type": "string"
                },
                "method": {
                    "type": "string"
                },
				"port": {
                    "type": "integer", 
					"minimum": 1024,
					"maximum": 65535
				},
				"tracePayloadOnNack" : {
					"type": "boolean",
					"default": false
				},
				"successStatus": {
                    "type": "integer", 
					"minimum": 200,
					"maximum": 599
				},
				"failureStatus": {
                    "type": "integer", 
					"minimum": 200,
					"maximum": 599
				}
            },
            "required": [
                "path",
				"port",
				"method"
            ],
            "title": "ReceiverConfig"
        }
    }
}
`
