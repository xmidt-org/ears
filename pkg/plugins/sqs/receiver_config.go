// SPDX-FileCopyrightText: 2025 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package sqs

import (
	"fmt"

	"github.com/xeipuuv/gojsonschema"
)

// WithDefaults returns a new config object that has all
// of the unset (nil) values filled in.
func (rc *ReceiverConfig) WithDefaults() ReceiverConfig {
	cfg := *rc
	if cfg.MaxNumberOfMessages == nil {
		cfg.MaxNumberOfMessages = DefaultReceiverConfig.MaxNumberOfMessages
	}
	if cfg.VisibilityTimeout == nil {
		cfg.VisibilityTimeout = DefaultReceiverConfig.VisibilityTimeout
	}
	if cfg.WaitTimeSeconds == nil {
		cfg.WaitTimeSeconds = DefaultReceiverConfig.WaitTimeSeconds
	}
	if cfg.AcknowledgeTimeout == nil {
		cfg.AcknowledgeTimeout = DefaultReceiverConfig.AcknowledgeTimeout
	}
	if cfg.NumRetries == nil {
		cfg.NumRetries = DefaultReceiverConfig.NumRetries
	}
	if cfg.ReceiverQueueDepth == nil {
		cfg.ReceiverQueueDepth = DefaultReceiverConfig.ReceiverQueueDepth
	}
	if cfg.ReceiverPoolSize == nil {
		cfg.ReceiverPoolSize = DefaultReceiverConfig.ReceiverPoolSize
	}
	if cfg.NeverDelete == nil {
		cfg.NeverDelete = DefaultReceiverConfig.NeverDelete
	}
	if cfg.TracePayloadOnNack == nil {
		cfg.TracePayloadOnNack = DefaultReceiverConfig.TracePayloadOnNack
	}
	if cfg.AWSRegion == "" {
		cfg.AWSRegion = DefaultReceiverConfig.AWSRegion
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
				"visibilityTimeout": {
                    "type": "integer", 
					"minimum": 1
				},
				"waitTimeSeconds": {
                    "type": "integer", 
					"minimum": 1
				},
				"acknowledgeTimeout": {
                    "type": "integer", 
					"minimum": 1,
					"maximum": 60
				},
				"numRetries": {
                    "type": "integer", 
					"minimum": 0,
					"maximum": 10
				},
				"receiverQueueDepth": {
                    "type": "integer", 
					"minimum": 0,
					"maximum": 1000
				},
				"receiverPoolSize": {
                    "type": "integer", 
					"minimum": 1,
					"maximum": 100
				},
				"neverDelete" : {
					"type": "boolean",
					"default": false
				},
				"tracePayloadOnNack" : {
					"type": "boolean",
					"default": false
				}
            },
            "required": [
                "queueUrl"
            ],
            "title": "ReceiverConfig"
        }
    }
}
`
