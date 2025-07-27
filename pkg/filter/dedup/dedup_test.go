// SPDX-FileCopyrightText: 2025 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package dedup_test

import (
	"context"
	"encoding/json"
	"reflect"
	"testing"

	"github.com/xmidt-org/ears/pkg/event"
	"github.com/xmidt-org/ears/pkg/filter/dedup"
	"github.com/xmidt-org/ears/pkg/tenant"
)

func TestFilterDedupBasic(t *testing.T) {
	ctx := context.Background()
	f, err := dedup.NewFilter(tenant.Id{AppId: "myapp", OrgId: "myorg"}, "dedup", "mydedup", dedup.Config{}, nil, nil)
	if err != nil {
		t.Fatalf("dedup test failed: %s\n", err.Error())
	}
	eventStr := `{"foo":"bar"}`
	var obj interface{}
	err = json.Unmarshal([]byte(eventStr), &obj)
	if err != nil {
		t.Fatalf("dedup test failed: %s\n", err.Error())
	}
	e, err := event.New(ctx, obj, event.FailOnNack(t))
	if err != nil {
		t.Fatalf("dedup test failed: %s\n", err.Error())
	}
	evts := f.Filter(e)
	if len(evts) != 1 {
		t.Fatalf("wrong number of deduped events: %d\n", len(evts))
	}
	expectedEventStr := `{"foo":"bar"}`
	var res interface{}
	err = json.Unmarshal([]byte(expectedEventStr), &res)
	if err != nil {
		t.Fatalf("dedup test failed: %s\n", err.Error())
	}
	if !reflect.DeepEqual(evts[0].Payload(), res) {
		t.Fatalf("wrong payload in deduped event: %v\n", evts[0])
	}
	evts = f.Filter(e)
	if len(evts) != 0 {
		t.Fatalf("dedup didn't filter events: %d\n", len(evts))
	}
}
