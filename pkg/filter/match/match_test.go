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

package match_test

import (
	"context"
	"encoding/json"
	"github.com/xmidt-org/ears/pkg/event"
	"github.com/xmidt-org/ears/pkg/filter/match"
	"github.com/xmidt-org/ears/pkg/filter/match/comparison"
	"github.com/xmidt-org/ears/pkg/tenant"
	"reflect"
	"testing"
)

func TestFilterMatchComparison1(t *testing.T) {
	ctx := context.Background()
	f, err := match.NewFilter(tenant.Id{AppId: "myapp", OrgId: "myorg"}, "match", "mymatch", match.Config{
		Comparison:    &comparison.Comparison{Equal: []map[string]interface{}{{"{.foo}": "bar"}, {"{.noo}": "nar"}}},
		PatternsLogic: "and",
		Matcher:       match.MatcherComparison,
		Mode:          match.ModeAllow,
	}, nil)
	if err != nil {
		t.Fatalf("comparison test failed: %s\n", err.Error())
	}
	eventStr := `{ "foo": "bar", "noo": "nar"}`
	var obj interface{}
	err = json.Unmarshal([]byte(eventStr), &obj)
	if err != nil {
		t.Fatalf("comparison test failed: %s\n", err.Error())
	}
	e, err := event.New(ctx, obj, event.FailOnNack(t))
	if err != nil {
		t.Fatalf("comparison test failed: %s\n", err.Error())
	}
	evts := f.Filter(e)
	if len(evts) != 1 {
		t.Fatalf("wrong number of compared events: %d\n", len(evts))
	}
	expectedEventStr := `{ "foo": "bar", "noo": "nar"}`
	var res interface{}
	err = json.Unmarshal([]byte(expectedEventStr), &res)
	if err != nil {
		t.Fatalf("comparison test failed: %s\n", err.Error())
	}
	if !reflect.DeepEqual(evts[0].Payload(), res) {
		pl, _ := json.MarshalIndent(evts[0].Payload(), "", "\t")
		t.Fatalf("wrong payload in compared event: %s\n", pl)
	}
}

func TestFilterMatchComparison1b(t *testing.T) {
	ctx := context.Background()
	f, err := match.NewFilter(tenant.Id{AppId: "myapp", OrgId: "myorg"}, "match", "mymatch", match.Config{
		Comparison:    &comparison.Comparison{NotEqual: []map[string]interface{}{{"{.foo}": "bar"}, {"{.noo}": "nar"}}},
		PatternsLogic: "and",
		Matcher:       match.MatcherComparison,
		Mode:          match.ModeAllow,
	}, nil)
	if err != nil {
		t.Fatalf("comparison test failed: %s\n", err.Error())
	}
	eventStr := `{ "foo": "bay", "noo": "nay"}`
	var obj interface{}
	err = json.Unmarshal([]byte(eventStr), &obj)
	if err != nil {
		t.Fatalf("comparison test failed: %s\n", err.Error())
	}
	e, err := event.New(ctx, obj, event.FailOnNack(t))
	if err != nil {
		t.Fatalf("comparison test failed: %s\n", err.Error())
	}
	evts := f.Filter(e)
	if len(evts) != 1 {
		t.Fatalf("wrong number of compared events: %d\n", len(evts))
	}
	expectedEventStr := `{ "foo": "bay", "noo": "nay"}`
	var res interface{}
	err = json.Unmarshal([]byte(expectedEventStr), &res)
	if err != nil {
		t.Fatalf("comparison test failed: %s\n", err.Error())
	}
	if !reflect.DeepEqual(evts[0].Payload(), res) {
		pl, _ := json.MarshalIndent(evts[0].Payload(), "", "\t")
		t.Fatalf("wrong payload in compared event: %s\n", pl)
	}
}

func TestFilterMatchComparisonTree1(t *testing.T) {
	ctx := context.Background()
	f, err := match.NewFilter(tenant.Id{AppId: "myapp", OrgId: "myorg"}, "match", "mymatch", match.Config{
		ComparisonTree: &comparison.ComparisonTreeNode{Logic: "and", Comparison: &comparison.Comparison{Equal: []map[string]interface{}{{"{.foo}": "bar"}, {"{.noo}": "nar"}}}},
		Matcher:        match.MatcherComparison,
		Mode:           match.ModeAllow,
	}, nil)
	if err != nil {
		t.Fatalf("comparison test failed: %s\n", err.Error())
	}
	eventStr := `{ "foo": "bar", "noo": "nar"}`
	var obj interface{}
	err = json.Unmarshal([]byte(eventStr), &obj)
	if err != nil {
		t.Fatalf("comparison test failed: %s\n", err.Error())
	}
	e, err := event.New(ctx, obj, event.FailOnNack(t))
	if err != nil {
		t.Fatalf("comparison test failed: %s\n", err.Error())
	}
	evts := f.Filter(e)
	if len(evts) != 1 {
		t.Fatalf("wrong number of compared events: %d\n", len(evts))
	}
	expectedEventStr := `{ "foo": "bar", "noo": "nar"}`
	var res interface{}
	err = json.Unmarshal([]byte(expectedEventStr), &res)
	if err != nil {
		t.Fatalf("comparison test failed: %s\n", err.Error())
	}
	if !reflect.DeepEqual(evts[0].Payload(), res) {
		pl, _ := json.MarshalIndent(evts[0].Payload(), "", "\t")
		t.Fatalf("wrong payload in compared event: %s\n", pl)
	}
}

func TestFilterMatchComparisonTree2(t *testing.T) {
	ctx := context.Background()
	tree2 := &comparison.ComparisonTreeNode{Logic: "or", Comparison: &comparison.Comparison{Equal: []map[string]interface{}{{"{.foo}": "bar"}, {"{.noo}": "nay"}}}}
	tree1 := &comparison.ComparisonTreeNode{Logic: "and", Comparison: &comparison.Comparison{Equal: []map[string]interface{}{{"{.foo}": "bar"}, {"{.noo}": "nar"}}}, ChildNodes: []*comparison.ComparisonTreeNode{tree2}}
	f, err := match.NewFilter(tenant.Id{AppId: "myapp", OrgId: "myorg"}, "match", "mymatch", match.Config{
		ComparisonTree: tree1,
		Matcher:        match.MatcherComparison,
		Mode:           match.ModeAllow,
	}, nil)
	if err != nil {
		t.Fatalf("comparison test failed: %s\n", err.Error())
	}
	eventStr := `{ "foo": "bar", "noo": "nar"}`
	var obj interface{}
	err = json.Unmarshal([]byte(eventStr), &obj)
	if err != nil {
		t.Fatalf("comparison test failed: %s\n", err.Error())
	}
	e, err := event.New(ctx, obj, event.FailOnNack(t))
	if err != nil {
		t.Fatalf("comparison test failed: %s\n", err.Error())
	}
	evts := f.Filter(e)
	if len(evts) != 1 {
		t.Fatalf("wrong number of compared events: %d\n", len(evts))
	}
	expectedEventStr := `{ "foo": "bar", "noo": "nar"}`
	var res interface{}
	err = json.Unmarshal([]byte(expectedEventStr), &res)
	if err != nil {
		t.Fatalf("comparison test failed: %s\n", err.Error())
	}
	if !reflect.DeepEqual(evts[0].Payload(), res) {
		pl, _ := json.MarshalIndent(evts[0].Payload(), "", "\t")
		t.Fatalf("wrong payload in compared event: %s\n", pl)
	}
}

func TestFilterMatchComparison2(t *testing.T) {
	ctx := context.Background()
	f, err := match.NewFilter(tenant.Id{AppId: "myapp", OrgId: "myorg"}, "match", "mymatch", match.Config{
		Comparison:    &comparison.Comparison{Equal: []map[string]interface{}{{"{.foo}": "bar"}, {"{.noo}": "narr"}}},
		PatternsLogic: "and",
		Matcher:       match.MatcherComparison,
		Mode:          match.ModeAllow,
	}, nil)
	if err != nil {
		t.Fatalf("comparison test failed: %s\n", err.Error())
	}
	eventStr := `{ "foo": "bar", "noo": "nar"}`
	var obj interface{}
	err = json.Unmarshal([]byte(eventStr), &obj)
	if err != nil {
		t.Fatalf("comparison test failed: %s\n", err.Error())
	}
	e, err := event.New(ctx, obj, event.FailOnNack(t))
	if err != nil {
		t.Fatalf("comparison test failed: %s\n", err.Error())
	}
	evts := f.Filter(e)
	if len(evts) != 0 {
		t.Fatalf("wrong number of compared events: %d\n", len(evts))
	}
}

func TestFilterMatchComparison3(t *testing.T) {
	ctx := context.Background()
	f, err := match.NewFilter(tenant.Id{AppId: "myapp", OrgId: "myorg"}, "match", "mymatch", match.Config{
		Comparison:    &comparison.Comparison{Equal: []map[string]interface{}{{"{.foo}": "bar"}, {"{.noo}": "narr"}}},
		PatternsLogic: "or",
		Matcher:       match.MatcherComparison,
		Mode:          match.ModeAllow,
	}, nil)
	if err != nil {
		t.Fatalf("comparison test failed: %s\n", err.Error())
	}
	eventStr := `{ "foo": "bar", "noo": "nar"}`
	var obj interface{}
	err = json.Unmarshal([]byte(eventStr), &obj)
	if err != nil {
		t.Fatalf("comparison test failed: %s\n", err.Error())
	}
	e, err := event.New(ctx, obj, event.FailOnNack(t))
	if err != nil {
		t.Fatalf("comparison test failed: %s\n", err.Error())
	}
	evts := f.Filter(e)
	if len(evts) != 1 {
		t.Fatalf("wrong number of compared events: %d\n", len(evts))
	}
	expectedEventStr := `{ "foo": "bar", "noo": "nar"}`
	var res interface{}
	err = json.Unmarshal([]byte(expectedEventStr), &res)
	if err != nil {
		t.Fatalf("comparison test failed: %s\n", err.Error())
	}
	if !reflect.DeepEqual(evts[0].Payload(), res) {
		pl, _ := json.MarshalIndent(evts[0].Payload(), "", "\t")
		t.Fatalf("wrong payload in compared event: %s\n", pl)
	}
}
