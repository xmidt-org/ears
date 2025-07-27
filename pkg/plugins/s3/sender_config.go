// SPDX-FileCopyrightText: 2025 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package s3

import (
	"fmt"

	"github.com/xeipuuv/gojsonschema"
)

// WithDefaults
func (sc SenderConfig) WithDefaults() SenderConfig {
	cfg := sc
	if cfg.AWSRegion == "" {
		cfg.AWSRegion = DefaultSenderConfig.AWSRegion
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
                "fileName": {
                    "type": "string"
                },
                "filePath": {
                    "type": "string"
                }
            },
            "required": [
                "bucket", "awsRegion"
            ],
            "title": "SenderConfig"
        }
    }
}
`
