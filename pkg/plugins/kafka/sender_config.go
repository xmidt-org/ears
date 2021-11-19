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

package kafka

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
	if cfg.Partition == nil {
		cfg.Partition = DefaultSenderConfig.Partition
	}
	if cfg.PartitionPath == "" {
		cfg.PartitionPath = DefaultSenderConfig.PartitionPath
	}
	if cfg.ChannelBufferSize == nil {
		cfg.ChannelBufferSize = DefaultSenderConfig.ChannelBufferSize
	}
	if cfg.Version == "" {
		cfg.Version = DefaultReceiverConfig.Version
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
                "brokers": {
                    "type": "string"
                },
                "topic": {
                    "type": "string"
                },
                "partition": {
                    "type": "integer"
                },
                "partitionPath": {
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
                "brokers", "topic"
            ],
            "title": "SenderConfig"
        }
    }
}
`
