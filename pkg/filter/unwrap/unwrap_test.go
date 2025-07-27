// SPDX-FileCopyrightText: 2025 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package unwrap_test

import (
	"context"
	"encoding/json"
	"reflect"
	"testing"

	"github.com/xmidt-org/ears/pkg/event"
	"github.com/xmidt-org/ears/pkg/filter/unwrap"
	"github.com/xmidt-org/ears/pkg/tenant"
)

func TestFilterUnwrapBasic(t *testing.T) {
	ctx := context.Background()
	f, err := unwrap.NewFilter(tenant.Id{AppId: "myapp", OrgId: "myorg"}, "unwrap", "myunwrap", unwrap.Config{
		Path: ".foo",
	}, nil, nil)
	if err != nil {
		t.Fatalf("unwrap test failed: %s\n", err.Error())
	}
	eventStr := `{ "foo": { "foo": "bar"}}`
	var obj interface{}
	err = json.Unmarshal([]byte(eventStr), &obj)
	if err != nil {
		t.Fatalf("unwrap test failed: %s\n", err.Error())
	}
	e, err := event.New(ctx, obj, event.FailOnNack(t))
	if err != nil {
		t.Fatalf("unwrap test failed: %s\n", err.Error())
	}
	evts := f.Filter(e)
	if len(evts) != 1 {
		t.Fatalf("wrong number of unwrapped events: %d\n", len(evts))
	}
	expectedEventStr := `{"foo":"bar"}`
	var res interface{}
	err = json.Unmarshal([]byte(expectedEventStr), &res)
	if err != nil {
		t.Fatalf("unwrap test failed: %s\n", err.Error())
	}
	if !reflect.DeepEqual(evts[0].Payload(), res) {
		pl, _ := json.MarshalIndent(evts[0].Payload(), "", "\t")
		t.Fatalf("wrong payload in unwrapped event: %s\n", pl)
	}
}
