// SPDX-FileCopyrightText: 2025 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

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
	if cfg.StartingTimestamp == nil {
		cfg.StartingTimestamp = DefaultReceiverConfig.StartingTimestamp
	}
	if cfg.EmptyStreamWaitSeconds == nil {
		cfg.EmptyStreamWaitSeconds = DefaultReceiverConfig.EmptyStreamWaitSeconds
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
	if rc.StreamArn == "" && rc.StreamName == "" {
		return fmt.Errorf("must provide either stream name or stream arn")
	}
	if rc.StreamArn != "" && rc.StreamName != "" {
		return fmt.Errorf("must provide either stream name or stream arn but not both")
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
                "streamArn": {
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
				"startingTimestamp": {
                    "type": "integer"
				},
				"startingSequenceNumber": {
                    "type": "string"
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
				},
				"emptyStreamWaitSeconds" : {
                    "type": "integer", 
					"minimum": 0
				}
            },
            "required": [
            ],
            "title": "ReceiverConfig"
        }
    }
}
`
