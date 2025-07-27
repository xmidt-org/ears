// SPDX-FileCopyrightText: 2025 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package merge_test

import (
	"context"
	"encoding/json"
	"reflect"
	"testing"

	"github.com/xmidt-org/ears/pkg/event"
	"github.com/xmidt-org/ears/pkg/filter/merge"
	"github.com/xmidt-org/ears/pkg/tenant"
)

func TestFilterMergeBasic(t *testing.T) {
	ctx := context.Background()
	f, err := merge.NewFilter(tenant.Id{AppId: "myapp", OrgId: "myorg"}, "merge", "mymerge", merge.Config{
		FromPath: ".foo",
		ToPath:   ".bar",
	}, nil, nil)
	if err != nil {
		t.Fatalf("merge test failed: %s\n", err.Error())
	}
	eventStr := `{"foo": ["a", "b"], "bar": ["b", "c"]}`
	var obj interface{}
	err = json.Unmarshal([]byte(eventStr), &obj)
	if err != nil {
		t.Fatalf("merge test failed: %s\n", err.Error())
	}
	e, err := event.New(ctx, obj, event.FailOnNack(t))
	if err != nil {
		t.Fatalf("merge test failed: %s\n", err.Error())
	}
	evts := f.Filter(e)
	if len(evts) != 1 {
		t.Fatalf("wrong number of merged events: %d\n", len(evts))
	}
	expectedEventStr := `{"foo": ["a", "b"], "bar": ["a", "b", "b", "c"]}`
	var res interface{}
	err = json.Unmarshal([]byte(expectedEventStr), &res)
	if err != nil {
		t.Fatalf("merge test failed: %s\n", err.Error())
	}
	if !reflect.DeepEqual(evts[0].Payload(), res) {
		pl, _ := json.MarshalIndent(evts[0].Payload(), "", "\t")
		t.Fatalf("wrong payload in merged event: %s\n", pl)
	}
}
