// Copyright 2021 Comcast Cable Communications Management, LLC
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

package validate_test

import (
	"context"
	"encoding/json"
	"github.com/xmidt-org/ears/pkg/event"
	"github.com/xmidt-org/ears/pkg/filter/validate"
	"github.com/xmidt-org/ears/pkg/tenant"
	"reflect"
	"testing"
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
	}, nil)
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
	}, nil)
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
