// SPDX-FileCopyrightText: 2025 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package js_test

import (
	"context"
	"encoding/json"
	"reflect"
	"testing"

	"github.com/xmidt-org/ears/pkg/event"
	"github.com/xmidt-org/ears/pkg/filter/js"
	"github.com/xmidt-org/ears/pkg/tenant"
)

func TestFilterJsBasic(t *testing.T) {
	ctx := context.Background()
	f, err := js.NewFilter(tenant.Id{AppId: "myapp", OrgId: "myorg"}, "js", "myjs", js.Config{
		Source: `var event = {};
        event.payload = _.event.payload;
        event.metadata =  _.event.metadata;
        event.payload.tag = "ears";
        return event;`,
	}, nil, nil)
	if err != nil {
		t.Fatalf("js test failed: %s\n", err.Error())
	}
	eventStr := `{ "foo": "bar"}`
	var obj interface{}
	err = json.Unmarshal([]byte(eventStr), &obj)
	if err != nil {
		t.Fatalf("js test failed: %s\n", err.Error())
	}
	e, err := event.New(ctx, obj, event.FailOnNack(t))
	if err != nil {
		t.Fatalf("js test failed: %s\n", err.Error())
	}
	evts := f.Filter(e)
	if len(evts) != 1 {
		t.Fatalf("wrong number of jsed events: %d\n", len(evts))
	}
	expectedEventStr := `{"foo":"bar", "tag":"ears"}`
	var res interface{}
	err = json.Unmarshal([]byte(expectedEventStr), &res)
	if err != nil {
		t.Fatalf("js test failed: %s\n", err.Error())
	}
	if !reflect.DeepEqual(evts[0].Payload(), res) {
		pl, _ := json.MarshalIndent(evts[0].Payload(), "", "\t")
		t.Fatalf("wrong payload in jsed event: %s\n", pl)
	}
}
