// SPDX-FileCopyrightText: 2025 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package block_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/xmidt-org/ears/pkg/event"
	"github.com/xmidt-org/ears/pkg/filter/block"
	"github.com/xmidt-org/ears/pkg/tenant"
)

func TestFilterPassBasic(t *testing.T) {
	ctx := context.Background()
	f, err := block.NewFilter(tenant.Id{AppId: "myapp", OrgId: "myorg"}, "block", "mymblock", block.Config{}, nil, nil)
	if err != nil {
		t.Fatalf("block test failed: %s\n", err.Error())
	}
	eventStr := `{ "foo": "bar"}`
	var obj interface{}
	err = json.Unmarshal([]byte(eventStr), &obj)
	if err != nil {
		t.Fatalf("block test failed: %s\n", err.Error())
	}
	e, err := event.New(ctx, obj, event.FailOnNack(t))
	if err != nil {
		t.Fatalf("block test failed: %s\n", err.Error())
	}
	evts := f.Filter(e)
	if len(evts) != 0 {
		t.Fatalf("wrong number of blocked events: %d\n", len(evts))
	}
}
