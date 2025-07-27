// SPDX-FileCopyrightText: 2025 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package batch_test

import (
	"context"
	"encoding/json"
	"reflect"
	"testing"

	"github.com/xmidt-org/ears/pkg/event"
	"github.com/xmidt-org/ears/pkg/filter/batch"
	"github.com/xmidt-org/ears/pkg/tenant"
	"github.com/xorcare/pointer"
)

func TestFilterBatchBasic(t *testing.T) {
	ctx := context.Background()
	f, err := batch.NewFilter(tenant.Id{AppId: "myapp", OrgId: "myorg"}, "batch", "mybatch", batch.Config{
		BatchSize: pointer.Int(2),
	}, nil, nil)
	if err != nil {
		t.Fatalf("batch test failed: %s\n", err.Error())
	}
	eventStr := `{ "foo": "bar"}`
	var obj interface{}
	err = json.Unmarshal([]byte(eventStr), &obj)
	if err != nil {
		t.Fatalf("batch test failed: %s\n", err.Error())
	}
	e, err := event.New(ctx, obj, event.FailOnNack(t))
	if err != nil {
		t.Fatalf("batch test failed: %s\n", err.Error())
	}
	evts := f.Filter(e)
	if len(evts) != 0 {
		t.Fatalf("wrong number of batched events: %d\n", len(evts))
	}
	evts = f.Filter(e)
	if len(evts) != 1 {
		t.Fatalf("wrong number of batched events: %d\n", len(evts))
	}
	expectedEventStr := `[{"foo":"bar"},{"foo":"bar"}]`
	var res interface{}
	err = json.Unmarshal([]byte(expectedEventStr), &res)
	if err != nil {
		t.Fatalf("batch test failed: %s\n", err.Error())
	}
	if !reflect.DeepEqual(evts[0].Payload(), res) {
		pl, _ := json.MarshalIndent(evts[0].Payload(), "", "\t")
		t.Fatalf("wrong payload in batched event: %s\n", pl)
	}
}
