// SPDX-FileCopyrightText: 2025 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package kafka

import (
	"fmt"

	"github.com/xeipuuv/gojsonschema"
)

// WithDefaults returns a new config object that has all
// of the unset (nil) values filled in.
func (rc *ReceiverConfig) WithDefaults() ReceiverConfig {
	cfg := *rc
	if cfg.Brokers == "" {
		cfg.Brokers = DefaultReceiverConfig.Brokers
	}
	if cfg.Topic == "" {
		cfg.Topic = DefaultReceiverConfig.Topic
	}
	if cfg.GroupId == "" {
		cfg.GroupId = DefaultReceiverConfig.GroupId
	}
	if cfg.CommitInterval == nil {
		cfg.CommitInterval = DefaultReceiverConfig.CommitInterval
	}
	if cfg.Version == "" {
		cfg.Version = DefaultReceiverConfig.Version
	}
	if cfg.ChannelBufferSize == nil {
		cfg.ChannelBufferSize = DefaultSenderConfig.ChannelBufferSize
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
                "brokers": {
                    "type": "string"
                },
                "topic": {
                    "type": "string"
                },
                "groupId": {
                    "type": "string"
                },
                "username": {
                    "type": "string"
                },
                "password": {
                    "type": "string"
                },
                "caCert": {
                    "type": "string"
                },
                "accessCert": {
                    "type": "string"
                },
                "accessKey": {
                    "type": "string"
                },
                "version": {
                    "type": "string"
                },
                "commitInterval": {
                    "type": "integer"
                },
                "channelBufferSize": {
                    "type": "integer"
                },
				"tracePayloadOnNack" : {
					"type": "boolean",
					"default": false
				}
            },
            "required": [
                "brokers", "topic", "groupId"
            ],
            "title": "ReceiverConfig"
        }
    }
}
`
