// SPDX-FileCopyrightText: 2025 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package validate_test

import (
	"context"
	"encoding/json"
	"reflect"
	"testing"

	"github.com/xmidt-org/ears/pkg/event"
	"github.com/xmidt-org/ears/pkg/filter/validate"
	"github.com/xmidt-org/ears/pkg/tenant"
)

func TestFilterValidateBasic(t *testing.T) {
	ctx := context.Background()
	myschema := `{
			"$schema": "http://json-schema.org/draft-06/schema#",
			"$ref": "#/definitions/MyEvent",
			"definitions": {
				"MyEvent": {
					"type": "object",
					"additionalProperties": false,
					"properties": {
						"foo": {
							"type": "string"
						}
					},
					"required": [
					"foo"
					],
					"title": "MyEvent"
				}
			}
		}`
	var sobj interface{}
	json.Unmarshal([]byte(myschema), &sobj)
	f, err := validate.NewFilter(tenant.Id{AppId: "myapp", OrgId: "myorg"}, "validate", "myvalidate", validate.Config{
		Path:   ".",
		Schema: sobj,
	}, nil, nil)
	if err != nil {
		t.Fatalf("validate test failed: %s\n", err.Error())
	}
	eventStr := `{"foo":"bar"}`
	var obj interface{}
	err = json.Unmarshal([]byte(eventStr), &obj)
	if err != nil {
		t.Fatalf("validate test failed: %s\n", err.Error())
	}
	e, err := event.New(ctx, obj, event.FailOnNack(t))
	if err != nil {
		t.Fatalf("validate test failed: %s\n", err.Error())
	}
	evts := f.Filter(e)
	if len(evts) != 1 {
		t.Fatalf("wrong number of validate events: %d\n", len(evts))
	}
	expectedEventStr := `{"foo":"bar"}`
	var res interface{}
	err = json.Unmarshal([]byte(expectedEventStr), &res)
	if err != nil {
		t.Fatalf("validate test failed: %s\n", err.Error())
	}
	if !reflect.DeepEqual(evts[0].Payload(), res) {
		pl, _ := json.MarshalIndent(evts[0].Payload(), "", "\t")
		t.Fatalf("wrong payload in validate event: %s\n", pl)
	}
}

func TestFilterValidateFilter(t *testing.T) {
	ctx := context.Background()
	myschema := `{
			"$schema": "http://json-schema.org/draft-06/schema#",
			"$ref": "#/definitions/MyEvent",
			"definitions": {
				"MyEvent": {
					"type": "object",
					"additionalProperties": false,
					"properties": {
						"foo": {
							"type": "string"
						}
					},
					"required": [
					"foo"
					],
					"title": "MyEvent"
				}
			}
		}`
	var sobj interface{}
	json.Unmarshal([]byte(myschema), &sobj)
	f, err := validate.NewFilter(tenant.Id{AppId: "myapp", OrgId: "myorg"}, "validate", "myvalidate", validate.Config{
		Path:   ".",
		Schema: sobj,
	}, nil, nil)
	if err != nil {
		t.Fatalf("validate test failed: %s\n", err.Error())
	}
	eventStr := `{"foo":"bar","baz":"whaz"}`
	var obj interface{}
	err = json.Unmarshal([]byte(eventStr), &obj)
	if err != nil {
		t.Fatalf("validate test failed: %s\n", err.Error())
	}
	e, err := event.New(ctx, obj, event.FailOnNack(t))
	if err != nil {
		t.Fatalf("validate test failed: %s\n", err.Error())
	}
	evts := f.Filter(e)
	if len(evts) != 0 {
		t.Fatalf("wrong number of validate events: %d\n", len(evts))
	}
}
