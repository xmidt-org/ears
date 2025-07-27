// SPDX-FileCopyrightText: 2025 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package sqs

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
	if cfg.DelaySeconds == nil {
		cfg.DelaySeconds = DefaultSenderConfig.DelaySeconds
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
                "queueUrl": {
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
					"maximum": 10
				},
				"sendTimeout": {
                    "type": "integer", 
					"minimum": 1,
					"maximum": 60
				},
				"delaySeconds": {
                    "type": "integer", 
					"minimum": 0,
					"maximum": 3600
				}
            },
            "required": [
                "queueUrl"
            ],
            "title": "SenderConfig"
        }
    }
}
`
