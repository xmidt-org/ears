// SPDX-FileCopyrightText: 2025 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package gears

import (
	"fmt"

	"github.com/xeipuuv/gojsonschema"
)

// WithDefaults
func (sc SenderConfig) WithDefaults() SenderConfig {
	cfg := sc
	if cfg.Brokers == "" {
		cfg.Brokers = DefaultSenderConfig.Brokers
	}
	if cfg.Topic == "" {
		cfg.Topic = DefaultSenderConfig.Topic
	}
	if cfg.Location == nil {
		cfg.Location = DefaultSenderConfig.Location
	}
	if cfg.App == "" {
		cfg.App = DefaultSenderConfig.App
	}
	if cfg.Partner == "" {
		cfg.Partner = DefaultSenderConfig.Partner
	}
	if cfg.ChannelBufferSize == nil {
		cfg.ChannelBufferSize = DefaultSenderConfig.ChannelBufferSize
	}
	if cfg.Version == "" {
		cfg.Version = DefaultSenderConfig.Version
	}
	if cfg.SenderPoolSize == nil {
		cfg.SenderPoolSize = DefaultSenderConfig.SenderPoolSize
	}
	if cfg.DynamicMetricLabels == nil {
		cfg.DynamicMetricLabels = DefaultSenderConfig.DynamicMetricLabels
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
                "clusters": {
					"type": "object"
				},
                "activeClusters": {
                    "type": "string"
                },
                "brokers": {
                    "type": "string"
                },
                "topic": {
                    "type": "string"
                },
                "location": {
                    "type": ["string", "array"]
                },
                "app": {
                    "type": "string"
                },
                "uses": {
                    "type": "string"
                },
                "enveloped": {
                    "type": "boolean"
                },
                "partner": {
                    "type": "string"
                },
                "metadata": {
                    "type": "boolean"
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
                "channelBufferSize": {
                    "type": "integer"
                },
				"senderPoolSize": {
                    "type": "integer", 
					"minimum": 1,
					"maximum": 100
				},
				"dynamicMetricLabel": {
                    "type": "array"
				},
                "compressionMethod": {
                    "type": "string"
                },
                "compressionLevel": {
                    "type": "integer"
                }
            },
            "required": [
                "topic"
            ],
            "title": "SenderConfig"
        }
    }
}
`
