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

package debug

import (
	"fmt"

	"github.com/xeipuuv/gojsonschema"
)

// WithDefaults returns a new config object that has all
// of the unset (nil) values filled in.
func (rc *ReceiverConfig) WithDefaults() ReceiverConfig {
	cfg := *rc
	if cfg.IntervalMs == nil {
		cfg.IntervalMs = DefaultReceiverConfig.IntervalMs
	}
	if cfg.Rounds == nil {
		cfg.Rounds = DefaultReceiverConfig.Rounds
	}
	if cfg.Payload == nil {
		cfg.Payload = DefaultReceiverConfig.Payload
	}
	if cfg.MaxHistory == nil {
		cfg.MaxHistory = DefaultReceiverConfig.MaxHistory
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

	// NOTE:  When `Min` was comparing against values other than 0,
	// this validation library just didn't work.  It seems it doesn't
	// work with pointers to ints

	// r := *rc

	// return validation.ValidateStruct(&r,
	// 	validation.Field(&r.IntervalMs,
	// 		validation.NotNil,
	// 		validation.Min(*minReceiverConfig.IntervalMs),
	// 	),
	// 	validation.Field(&r.Rounds,
	// 		validation.NotNil,
	// 		validation.Min(*minReceiverConfig.Rounds),
	// 	),
	// 	validation.Field(&r.MaxHistory,
	// 		validation.NotNil,
	// 		validation.Min(*minReceiverConfig.MaxHistory),
	// 	),
	// )
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
                "intervalMs": {
                    "type": "integer",
					"minimum": 1
                },
                "rounds": {
                    "type": "integer",
					"minimum": -1
                },
                "payload": {
                    "type": ["string", "object", "array"]
                },
                "maxHistory": {
                    "type": "integer",
					"minimum": 0
                },
				"tracePayloadOnNack" : {
					"type": "boolean",
					"default": false
				}
            },
            "required": [
                "intervalMs",
                "maxHistory",
                "rounds"
            ],
            "title": "ReceiverConfig"
        }
    }
}
`
