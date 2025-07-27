// SPDX-FileCopyrightText: 2025 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package encode_test

import (
	"context"
	"encoding/json"
	"reflect"
	"testing"

	"github.com/xmidt-org/ears/pkg/event"
	"github.com/xmidt-org/ears/pkg/filter/encode"
	"github.com/xmidt-org/ears/pkg/tenant"
)

func TestFilterEncodeBasic(t *testing.T) {
	ctx := context.Background()
	f, err := encode.NewFilter(tenant.Id{AppId: "myapp", OrgId: "myorg"}, "encode", "myencode", encode.Config{
		FromPath: ".",
		ToPath:   ".value",
	}, nil, nil)
	if err != nil {
		t.Fatalf("encode test failed: %s\n", err.Error())
	}
	eventStr := `{"foo":"bar"}`
	var obj interface{}
	err = json.Unmarshal([]byte(eventStr), &obj)
	if err != nil {
		t.Fatalf("encode test failed: %s\n", err.Error())
	}
	e, err := event.New(ctx, obj, event.FailOnNack(t))
	if err != nil {
		t.Fatalf("encode test failed: %s\n", err.Error())
	}
	evts := f.Filter(e)
	if len(evts) != 1 {
		t.Fatalf("wrong number of encoded events: %d\n", len(evts))
	}
	expectedEventStr := `{ "foo" : "bar", "value": "eyJmb28iOiJiYXIifQ=="}`
	var res interface{}
	err = json.Unmarshal([]byte(expectedEventStr), &res)
	if err != nil {
		t.Fatalf("encode test failed: %s\n", err.Error())
	}
	if !reflect.DeepEqual(evts[0].Payload(), res) {
		pl, _ := json.MarshalIndent(evts[0].Payload(), "", "\t")
		t.Fatalf("wrong payload in encoded event: %s\n", pl)
	}
}
