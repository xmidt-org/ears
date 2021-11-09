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

package s3

import (
	"fmt"
	"github.com/xeipuuv/gojsonschema"
)

// WithDefaults
func (sc SenderConfig) WithDefaults() SenderConfig {
	cfg := sc
	if cfg.Region == "" {
		cfg.Region = DefaultSenderConfig.Region
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
                "bucket": {
                    "type": "string"
                },
                "path": {
                    "type": "string"
                },
                "region": {
                    "type": "string"
                },
                "fileName": {
                    "type": "string"
                },
                "filePath": {
                    "type": "string"
                }
            },
            "required": [
                "bucket", "region"
            ],
            "title": "SenderConfig"
        }
    }
}
`
