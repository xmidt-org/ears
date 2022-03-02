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

package kinesis

import (
	"fmt"

	"github.com/xeipuuv/gojsonschema"
)

// WithDefaults returns a new config object that has all
// of the unset (nil) values filled in.
func (rc *ReceiverConfig) WithDefaults() ReceiverConfig {
	cfg := *rc
	if cfg.AcknowledgeTimeout == nil {
		cfg.AcknowledgeTimeout = DefaultReceiverConfig.AcknowledgeTimeout
	}
	if cfg.ShardIteratorType == "" {
		cfg.ShardIteratorType = DefaultReceiverConfig.ShardIteratorType
	}
	if cfg.TracePayloadOnNack == nil {
		cfg.TracePayloadOnNack = DefaultReceiverConfig.TracePayloadOnNack
	}
	if cfg.EnhancedFanOut == nil {
		cfg.EnhancedFanOut = DefaultReceiverConfig.EnhancedFanOut
	}
	if cfg.AWSRegion == "" {
		cfg.AWSRegion = DefaultReceiverConfig.AWSRegion
	}
	if cfg.UseCheckpoint == nil {
		cfg.UseCheckpoint = DefaultReceiverConfig.UseCheckpoint
	}
	if cfg.MaxCheckpointAgeSeconds == nil {
		cfg.MaxCheckpointAgeSeconds = DefaultReceiverConfig.MaxCheckpointAgeSeconds
	}
	if cfg.UseShardMonitor == nil {
		cfg.UseShardMonitor = DefaultReceiverConfig.UseShardMonitor
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
	if *rc.EnhancedFanOut && rc.ConsumerName == "" {
		return fmt.Errorf("must provide consumer name with enhanced fan-out option")
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
                "streamName": {
                    "type": "string"
                },
                "consumerName": {
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
                "shardIteratorType": {
                    "type": "string"
                },
				"acknowledgeTimeout": {
                    "type": "integer", 
					"minimum": 1,
					"maximum": 60
				},
				"tracePayloadOnNack" : {
					"type": "boolean",
					"default": false
				},
				"enhancedFanOut" : {
					"type": "boolean",
					"default": false
				},
				"useCheckpoint" : {
					"type": "boolean",
					"default": true
				},
				"useShardMonitor" : {
					"type": "boolean",
					"default": false
				},
				"maxCheckpointAgeSeconds" : {
                    "type": "integer", 
					"minimum": 0
				}
            },
            "required": [
                "streamName"
            ],
            "title": "ReceiverConfig"
        }
    }
}
`
