// SPDX-FileCopyrightText: 2025 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package kinesis

import (
	"fmt"

	"github.com/xeipuuv/gojsonschema"
)

// WithDefaults
func (sc SenderConfig) WithDefaults() SenderConfig {
	cfg := sc
	if cfg.MaxNumberOfMessages == nil {
		cfg.MaxNumberOfMessages = DefaultSenderConfig.MaxNumberOfMessages
	}
	if cfg.SendTimeout == nil {
		cfg.SendTimeout = DefaultSenderConfig.SendTimeout
	}
	if cfg.AWSRegion == "" {
		cfg.AWSRegion = DefaultReceiverConfig.AWSRegion
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
	if sc.StreamArn == "" && sc.StreamName == "" {
		return fmt.Errorf("must provide either stream name or stream arn")
	}
	if sc.StreamArn != "" && sc.StreamName != "" {
		return fmt.Errorf("must provide either stream name or stream arn but not both")
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
                "streamName": {
                    "type": "string"
                },
                "streamArn": {
                    "type": "string"
                },
                "partitionKey": {
                    "type": "string"
                },
                "partitionKeyPath": {
                    "type": "string"
                },
				"awsRoleARN": {
                    "type": "string"
				},
				"awsAccessKeyId": {
                    "type": "string"
				},
				"awsSecretAccessKey": {
                    "type": "string"
				},
				"awsRegion": {
                    "type": "string"
				},
				"maxNumberOfMessages": {
                    "type": "integer", 
					"minimum": 1,
					"maximum": 500
				},
				"sendTimeout": {
                    "type": "integer", 
					"minimum": 1,
					"maximum": 60
				}
            },
            "required": [
            ],
            "title": "SenderConfig"
        }
    }
}
`
