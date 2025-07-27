// SPDX-FileCopyrightText: 2025 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package transform_test

import (
	"context"
	"encoding/json"
	"reflect"
	"testing"

	"github.com/xmidt-org/ears/pkg/event"
	"github.com/xmidt-org/ears/pkg/filter/transform"
	"github.com/xmidt-org/ears/pkg/tenant"
)

func TestFilterTransformBasic(t *testing.T) {
	ctx := context.Background()
	f, err := transform.NewFilter(tenant.Id{AppId: "myapp", OrgId: "myorg"}, "transform", "mytransform", transform.Config{
		ToPath:         ".",
		Transformation: map[string]interface{}{"bar": "{.foo}"},
	}, nil, nil)
	if err != nil {
		t.Fatalf("transform test failed: %s\n", err.Error())
	}
	eventStr := `{"foo":"bar"}`
	var obj interface{}
	err = json.Unmarshal([]byte(eventStr), &obj)
	if err != nil {
		t.Fatalf("transform test failed: %s\n", err.Error())
	}
	e, err := event.New(ctx, obj, event.FailOnNack(t))
	if err != nil {
		t.Fatalf("transform test failed: %s\n", err.Error())
	}
	evts := f.Filter(e)
	if len(evts) != 1 {
		t.Fatalf("wrong number of transformed events: %d\n", len(evts))
	}
	expectedEventStr := `{"bar":"bar"}`
	var res interface{}
	err = json.Unmarshal([]byte(expectedEventStr), &res)
	if err != nil {
		t.Fatalf("transform test failed: %s\n", err.Error())
	}
	if !reflect.DeepEqual(evts[0].Payload(), res) {
		pl, _ := json.MarshalIndent(evts[0].Payload(), "", "\t")
		t.Fatalf("wrong payload in transfomred event: %s\n", pl)
	}
}

func TestFilterTransformArray(t *testing.T) {
	ctx := context.Background()
	f, err := transform.NewFilter(tenant.Id{AppId: "myapp", OrgId: "myorg"}, "transform", "mytransform", transform.Config{
		FromPath:       ".foo",
		ToPath:         ".foo",
		Transformation: map[string]interface{}{"bar": "{.foo}"},
	}, nil, nil)
	if err != nil {
		t.Fatalf("transform test failed: %s\n", err.Error())
	}
	eventStr := `{"foo":[{"foo":"bar"},{"foo":"bar"}]}`
	var obj interface{}
	err = json.Unmarshal([]byte(eventStr), &obj)
	if err != nil {
		t.Fatalf("transform test failed: %s\n", err.Error())
	}
	e, err := event.New(ctx, obj, event.FailOnNack(t))
	if err != nil {
		t.Fatalf("transform test failed: %s\n", err.Error())
	}
	evts := f.Filter(e)
	if len(evts) != 1 {
		t.Fatalf("wrong number of transformed events: %d\n", len(evts))
	}
	expectedEventStr := `{"foo":[{"bar":"bar"},{"bar":"bar"}]}`
	var res interface{}
	err = json.Unmarshal([]byte(expectedEventStr), &res)
	if err != nil {
		t.Fatalf("transform test failed: %s\n", err.Error())
	}
	if !reflect.DeepEqual(evts[0].Payload(), res) {
		pl, _ := json.MarshalIndent(evts[0].Payload(), "", "\t")
		t.Fatalf("wrong payload in transfomred event: %s\n", pl)
	}
}
