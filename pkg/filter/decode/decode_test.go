// SPDX-FileCopyrightText: 2025 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package decode_test

import (
	"context"
	"encoding/json"
	"reflect"
	"testing"

	"github.com/xmidt-org/ears/pkg/event"
	"github.com/xmidt-org/ears/pkg/filter/decode"
	"github.com/xmidt-org/ears/pkg/tenant"
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
