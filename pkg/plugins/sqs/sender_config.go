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

// WithDefaults
func (sc SenderConfig) WithDefaults() SenderConfig {
	cfg := sc
	if cfg.BatchSize == nil {
		cfg.BatchSize = DefaultSenderConfig.BatchSize
	}
	if cfg.SendTimeout == nil {
		cfg.SendTimeout = DefaultSenderConfig.SendTimeout
	}
	if cfg.DelaySeconds == nil {
		cfg.DelaySeconds = DefaultSenderConfig.DelaySeconds
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
				"batchSize": {
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
