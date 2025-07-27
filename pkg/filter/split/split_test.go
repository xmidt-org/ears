// SPDX-FileCopyrightText: 2025 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package split_test

import (
	"context"
	"encoding/json"
	"reflect"
	"testing"

	"github.com/xmidt-org/ears/pkg/event"
	"github.com/xmidt-org/ears/pkg/filter/split"
	"github.com/xmidt-org/ears/pkg/tenant"
)

func TestFilterSplitBasic(t *testing.T) {
	ctx := context.Background()
	f, err := split.NewFilter(tenant.Id{AppId: "myapp", OrgId: "myorg"}, "split", "mysplit", split.Config{
		Path: ".foo",
	}, nil, nil)
	if err != nil {
		t.Fatalf("split test failed: %s\n", err.Error())
	}
	eventStr := `{ "foo": [{"foo":"bar"},{"foo":"bar"}]}`
	var obj interface{}
	err = json.Unmarshal([]byte(eventStr), &obj)
	if err != nil {
		t.Fatalf("split test failed: %s\n", err.Error())
	}
	e, err := event.New(ctx, obj, event.FailOnNack(t))
	if err != nil {
		t.Fatalf("split test failed: %s\n", err.Error())
	}
	evts := f.Filter(e)
	if len(evts) != 2 {
		t.Fatalf("wrong number of splitted events: %d\n", len(evts))
	}
	expectedEventStr := `{"foo":"bar"}`
	var res interface{}
	err = json.Unmarshal([]byte(expectedEventStr), &res)
	if err != nil {
		t.Fatalf("split test failed: %s\n", err.Error())
	}
	if !reflect.DeepEqual(evts[0].Payload(), res) {
		pl, _ := json.MarshalIndent(evts[0].Payload(), "", "\t")
		t.Fatalf("wrong payload in splitted event: %s\n", pl)
	}
	if !reflect.DeepEqual(evts[0].Payload(), res) {
		pl, _ := json.MarshalIndent(evts[1].Payload(), "", "\t")
		t.Fatalf("wrong payload in splitted event: %s\n", pl)
	}
}
