// SPDX-FileCopyrightText: 2025 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package modify_test

import (
	"context"
	"encoding/json"
	"reflect"
	"testing"

	"github.com/xmidt-org/ears/pkg/event"
	"github.com/xmidt-org/ears/pkg/filter/modify"
	"github.com/xmidt-org/ears/pkg/tenant"
	"github.com/xorcare/pointer"
)

func TestFilterModifyUpper(t *testing.T) {
	ctx := context.Background()
	f, err := modify.NewFilter(tenant.Id{AppId: "myapp", OrgId: "myorg"}, "modify", "mymodify", modify.Config{
		Path:    ".foo",
		ToUpper: pointer.Bool(true),
	}, nil, nil)
	if err != nil {
		t.Fatalf("modify test failed: %s\n", err.Error())
	}
	eventStr := `{ "foo": "bar" }`
	var obj interface{}
	err = json.Unmarshal([]byte(eventStr), &obj)
	if err != nil {
		t.Fatalf("modify test failed: %s\n", err.Error())
	}
	e, err := event.New(ctx, obj, event.FailOnNack(t))
	if err != nil {
		t.Fatalf("modify test failed: %s\n", err.Error())
	}
	evts := f.Filter(e)
	if len(evts) != 1 {
		t.Fatalf("wrong number of events: %d\n", len(evts))
	}
	expectedEventStr := `{ "foo": "BAR" }`
	var res interface{}
	err = json.Unmarshal([]byte(expectedEventStr), &res)
	if err != nil {
		t.Fatalf("modify test failed: %s\n", err.Error())
	}
	if !reflect.DeepEqual(evts[0].Payload(), res) {
		pl, _ := json.MarshalIndent(evts[0].Payload(), "", "\t")
		t.Fatalf("wrong payload in modified event: %s\n", pl)
	}
}

func TestFilterModifyLower(t *testing.T) {
	ctx := context.Background()
	f, err := modify.NewFilter(tenant.Id{AppId: "myapp", OrgId: "myorg"}, "modify", "mymodify", modify.Config{
		Path:    ".foo",
		ToLower: pointer.Bool(true),
	}, nil, nil)
	if err != nil {
		t.Fatalf("modify test failed: %s\n", err.Error())
	}
	eventStr := `{ "foo": "BAR" }`
	var obj interface{}
	err = json.Unmarshal([]byte(eventStr), &obj)
	if err != nil {
		t.Fatalf("modify test failed: %s\n", err.Error())
	}
	e, err := event.New(ctx, obj, event.FailOnNack(t))
	if err != nil {
		t.Fatalf("modify test failed: %s\n", err.Error())
	}
	evts := f.Filter(e)
	if len(evts) != 1 {
		t.Fatalf("wrong number of events: %d\n", len(evts))
	}
	expectedEventStr := `{ "foo": "bar" }`
	var res interface{}
	err = json.Unmarshal([]byte(expectedEventStr), &res)
	if err != nil {
		t.Fatalf("modify test failed: %s\n", err.Error())
	}
	if !reflect.DeepEqual(evts[0].Payload(), res) {
		pl, _ := json.MarshalIndent(evts[0].Payload(), "", "\t")
		t.Fatalf("wrong payload in modified event: %s\n", pl)
	}
}

func TestFilterModifyMany(t *testing.T) {
	ctx := context.Background()
	f, err := modify.NewFilter(tenant.Id{AppId: "myapp", OrgId: "myorg"}, "modify", "mymodify", modify.Config{
		Paths:   []string{".foo", ".a"},
		ToLower: pointer.Bool(true),
	}, nil, nil)
	if err != nil {
		t.Fatalf("modify test failed: %s\n", err.Error())
	}
	eventStr := `{ "foo": "BAR", "a": "VALUE" }`
	var obj interface{}
	err = json.Unmarshal([]byte(eventStr), &obj)
	if err != nil {
		t.Fatalf("modify test failed: %s\n", err.Error())
	}
	e, err := event.New(ctx, obj, event.FailOnNack(t))
	if err != nil {
		t.Fatalf("modify test failed: %s\n", err.Error())
	}
	evts := f.Filter(e)
	if len(evts) != 1 {
		t.Fatalf("wrong number of events: %d\n", len(evts))
	}
	expectedEventStr := `{ "foo": "bar", "a": "value" }`
	var res interface{}
	err = json.Unmarshal([]byte(expectedEventStr), &res)
	if err != nil {
		t.Fatalf("modify test failed: %s\n", err.Error())
	}
	if !reflect.DeepEqual(evts[0].Payload(), res) {
		pl, _ := json.MarshalIndent(evts[0].Payload(), "", "\t")
		t.Fatalf("wrong payload in modified event: %s\n", pl)
	}
}

func TestFilterModifyDeadPath(t *testing.T) {
	ctx := context.Background()
	f, err := modify.NewFilter(tenant.Id{AppId: "myapp", OrgId: "myorg"}, "modify", "mymodify", modify.Config{
		Path:    ".foo.bar.baz",
		ToUpper: pointer.Bool(true),
	}, nil, nil)
	if err != nil {
		t.Fatalf("modify test failed: %s\n", err.Error())
	}
	eventStr := `{ "foo": "bar" }`
	var obj interface{}
	err = json.Unmarshal([]byte(eventStr), &obj)
	if err != nil {
		t.Fatalf("modify test failed: %s\n", err.Error())
	}
	e, err := event.New(ctx, obj, event.FailOnNack(t))
	if err != nil {
		t.Fatalf("modify test failed: %s\n", err.Error())
	}
	evts := f.Filter(e)
	if len(evts) != 1 {
		t.Fatalf("wrong number of events: %d\n", len(evts))
	}
	expectedEventStr := `{ "foo": "bar" }`
	var res interface{}
	err = json.Unmarshal([]byte(expectedEventStr), &res)
	if err != nil {
		t.Fatalf("modify test failed: %s\n", err.Error())
	}
	if !reflect.DeepEqual(evts[0].Payload(), res) {
		pl, _ := json.MarshalIndent(evts[0].Payload(), "", "\t")
		t.Fatalf("wrong payload in modified event: %s\n", pl)
	}
}
