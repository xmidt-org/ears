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

package decode_test

import (
	"context"
	"encoding/json"
	"github.com/xmidt-org/ears/pkg/event"
	"github.com/xmidt-org/ears/pkg/filter/decode"
	"github.com/xmidt-org/ears/pkg/tenant"
	"reflect"
	"testing"
)

func TestFilterDecodeBasic(t *testing.T) {
	ctx := context.Background()
	f, err := decode.NewFilter(tenant.Id{AppId: "myapp", OrgId: "myorg"}, "decode", "mydecode", decode.Config{
		FromPath: ".value",
		ToPath:   ".",
		Encoding: "base64",
	}, nil, nil)
	if err != nil {
		t.Fatalf("decode test failed: %s\n", err.Error())
	}
	eventStr := `{ "value": "eyJmb28iOiJiYXIifQ=="}`
	var obj interface{}
	err = json.Unmarshal([]byte(eventStr), &obj)
	if err != nil {
		t.Fatalf("decode test failed: %s\n", err.Error())
	}
	e, err := event.New(ctx, obj, event.FailOnNack(t))
	if err != nil {
		t.Fatalf("decode test failed: %s\n", err.Error())
	}
	evts := f.Filter(e)
	if len(evts) != 1 {
		t.Fatalf("wrong number of decoded events: %d\n", len(evts))
	}
	expectedEventStr := `{"foo":"bar"}`
	var res interface{}
	err = json.Unmarshal([]byte(expectedEventStr), &res)
	if err != nil {
		t.Fatalf("decode test failed: %s\n", err.Error())
	}
	if !reflect.DeepEqual(evts[0].Payload(), res) {
		pl, _ := json.MarshalIndent(evts[0].Payload(), "", "\t")
		t.Fatalf("wrong payload in decoded event: %s\n", pl)
	}
}

func TestFilterDecodeString(t *testing.T) {
	ctx := context.Background()
	f, err := decode.NewFilter(tenant.Id{AppId: "myapp", OrgId: "myorg"}, "decode", "mydecode", decode.Config{
		FromPath: ".value",
		ToPath:   ".value",
		Encoding: "string",
	}, nil, nil)
	if err != nil {
		t.Fatalf("decode test failed: %s\n", err.Error())
	}
	eventStr := `{ "value": "{\"foo\":\"bar\"}"}`
	var obj interface{}
	err = json.Unmarshal([]byte(eventStr), &obj)
	if err != nil {
		t.Fatalf("decode test failed: %s\n", err.Error())
	}
	e, err := event.New(ctx, obj, event.FailOnNack(t))
	if err != nil {
		t.Fatalf("decode test failed: %s\n", err.Error())
	}
	evts := f.Filter(e)
	if len(evts) != 1 {
		t.Fatalf("wrong number of decoded events: %d\n", len(evts))
	}
	expectedEventStr := `{"value":{"foo":"bar"}}`
	var res interface{}
	err = json.Unmarshal([]byte(expectedEventStr), &res)
	if err != nil {
		t.Fatalf("decode test failed: %s\n", err.Error())
	}
	if !reflect.DeepEqual(evts[0].Payload(), res) {
		pl, _ := json.MarshalIndent(evts[0].Payload(), "", "\t")
		t.Fatalf("wrong payload in decoded event: %s\n", pl)
	}
}
