package http

import (
	"fmt"
	"github.com/xeipuuv/gojsonschema"
)

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
                "path": {
                    "type": "string"
                },
                "method": {
                    "type": "string"
                },
				"port": {
                    "type": "integer", 
					"minimum": 1024,
					"maximum": 65535
				}
            },
            "required": [
                "path",
				"port",
				"method"
            ],
            "title": "ReceiverConfig"
        }
    }
}
`
