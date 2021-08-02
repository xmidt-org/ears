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
