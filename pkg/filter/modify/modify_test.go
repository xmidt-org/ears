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

package modify_test

import (
	"context"
	"encoding/json"
	"github.com/xmidt-org/ears/pkg/event"
	"github.com/xmidt-org/ears/pkg/filter/modify"
	"github.com/xmidt-org/ears/pkg/tenant"
	"github.com/xorcare/pointer"
	"reflect"
	"testing"
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
