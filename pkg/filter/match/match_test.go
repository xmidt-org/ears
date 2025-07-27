// SPDX-FileCopyrightText: 2025 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package match_test

import (
	"context"
	"encoding/json"
	"reflect"
	"testing"

	"github.com/xmidt-org/ears/pkg/event"
	"github.com/xmidt-org/ears/pkg/filter/match"
	"github.com/xmidt-org/ears/pkg/filter/match/comparison"
	"github.com/xmidt-org/ears/pkg/tenant"
)

func TestFilterMatchComparison1(t *testing.T) {
	ctx := context.Background()
	f, err := match.NewFilter(tenant.Id{AppId: "myapp", OrgId: "myorg"}, "match", "mymatch", match.Config{
		Comparison:    &comparison.Comparison{Equal: []map[string]interface{}{{"{.foo}": "bar"}, {"{.noo}": "nar"}}},
		PatternsLogic: "and",
		Matcher:       match.MatcherComparison,
		Mode:          match.ModeAllow,
	}, nil, nil)
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
	}, nil, nil)
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
	}, nil, nil)
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
	}, nil, nil)
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
	}, nil, nil)
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
	}, nil, nil)
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
