// SPDX-FileCopyrightText: 2025 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package mapping_test

import (
	"context"
	"encoding/json"
	"reflect"
	"testing"

	"github.com/xmidt-org/ears/pkg/event"
	"github.com/xmidt-org/ears/pkg/filter/mapping"
	"github.com/xmidt-org/ears/pkg/tenant"
)

func TestFilterMappingBasic(t *testing.T) {
	ctx := context.Background()
	f, err := mapping.NewFilter(tenant.Id{AppId: "myapp", OrgId: "myorg"}, "mapping", "mymapping", mapping.Config{
		Path:         ".foo",
		ArrayPath:    "",
		Map:          []mapping.FromTo{mapping.FromTo{From: "bar", To: true}},
		DefaultValue: nil,
	}, nil, nil)
	if err != nil {
		t.Fatalf("mapping test failed: %s\n", err.Error())
	}
	eventStr := `{"foo":"bar"}`
	var obj interface{}
	err = json.Unmarshal([]byte(eventStr), &obj)
	if err != nil {
		t.Fatalf("mapping test failed: %s\n", err.Error())
	}
	e, err := event.New(ctx, obj, event.FailOnNack(t))
	if err != nil {
		t.Fatalf("mapping test failed: %s\n", err.Error())
	}
	evts := f.Filter(e)
	if len(evts) != 1 {
		t.Fatalf("wrong number of mapped events: %d\n", len(evts))
	}
	expectedEventStr := `{"foo":true}`
	var res interface{}
	err = json.Unmarshal([]byte(expectedEventStr), &res)
	if err != nil {
		t.Fatalf("mapping test failed: %s\n", err.Error())
	}
	if !reflect.DeepEqual(evts[0].Payload(), res) {
		t.Fatalf("wrong payload in mapped event: %v\n", evts[0])
	}
}

func TestFilterMappingDefault(t *testing.T) {
	ctx := context.Background()
	f, err := mapping.NewFilter(tenant.Id{AppId: "myapp", OrgId: "myorg"}, "match", "mymatch", mapping.Config{
		Path:         ".foo",
		ArrayPath:    "",
		Map:          []mapping.FromTo{mapping.FromTo{From: "baz", To: true}},
		DefaultValue: "default",
	}, nil, nil)
	if err != nil {
		t.Fatalf("mapping test failed: %s\n", err.Error())
	}
	eventStr := `{"foo":"bar"}`
	var obj interface{}
	err = json.Unmarshal([]byte(eventStr), &obj)
	if err != nil {
		t.Fatalf("mapping test failed: %s\n", err.Error())
	}
	e, err := event.New(ctx, obj, event.FailOnNack(t))
	if err != nil {
		t.Fatalf("mapping test failed: %s\n", err.Error())
	}
	evts := f.Filter(e)
	if len(evts) != 1 {
		t.Fatalf("wrong number of mapped events: %d\n", len(evts))
	}
	expectedEventStr := `{"foo":"default"}`
	var res interface{}
	err = json.Unmarshal([]byte(expectedEventStr), &res)
	if err != nil {
		t.Fatalf("mapping test failed: %s\n", err.Error())
	}
	if !reflect.DeepEqual(evts[0].Payload(), res) {
		t.Fatalf("wrong payload in mapped event: %v\n", evts[0])
	}
}

func TestFilterMappingArray(t *testing.T) {
	ctx := context.Background()
	f, err := mapping.NewFilter(tenant.Id{AppId: "myapp", OrgId: "myorg"}, "match", "mymatch", mapping.Config{
		Path:         ".hello",
		ArrayPath:    ".foo",
		Map:          []mapping.FromTo{mapping.FromTo{From: "earth", To: "planet"}, mapping.FromTo{From: "mars", To: "planet"}},
		DefaultValue: nil,
	}, nil, nil)
	if err != nil {
		t.Fatalf("mapping test failed: %s\n", err.Error())
	}
	eventStr := `{"foo":[{"hello":"earth"},{"hello":"mars"}]}`
	var obj interface{}
	err = json.Unmarshal([]byte(eventStr), &obj)
	if err != nil {
		t.Fatalf("mapping test failed: %s\n", err.Error())
	}
	e, err := event.New(ctx, obj, event.FailOnNack(t))
	if err != nil {
		t.Fatalf("mapping test failed: %s\n", err.Error())
	}
	evts := f.Filter(e)
	if len(evts) != 1 {
		t.Fatalf("wrong number of mapped events: %d\n", len(evts))
	}
	expectedEventStr := `{"foo":[{"hello":"planet"},{"hello":"planet"}]}`
	var res interface{}
	err = json.Unmarshal([]byte(expectedEventStr), &res)
	if err != nil {
		t.Fatalf("mapping test failed: %s\n", err.Error())
	}
	if !reflect.DeepEqual(evts[0].Payload(), res) {
		t.Fatalf("wrong payload in mapped event: %v\n", evts[0])
	}
}
