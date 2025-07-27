// SPDX-FileCopyrightText: 2025 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package ttl_test

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/xmidt-org/ears/pkg/event"
	"github.com/xmidt-org/ears/pkg/filter/ttl"
	"github.com/xmidt-org/ears/pkg/tenant"
	"github.com/xorcare/pointer"
)

func TestFilterTtlBasic(t *testing.T) {
	ctx := context.Background()
	f, err := ttl.NewFilter(tenant.Id{AppId: "myapp", OrgId: "myorg"}, "ttl", "myttl", ttl.Config{
		Path:       ".ts",
		NanoFactor: pointer.Int(1),
		Ttl:        pointer.Int(300000),
	}, nil, nil)
	if err != nil {
		t.Fatalf("ttl test failed: %s\n", err.Error())
	}
	ts := time.Now().UnixNano()
	tsStr := fmt.Sprintf("%d", ts)
	eventStr := `{ "foo": "bar", "ts":` + tsStr + "}"
	var obj interface{}
	err = json.Unmarshal([]byte(eventStr), &obj)
	if err != nil {
		t.Fatalf("ttl test failed: %s\n", err.Error())
	}
	e, err := event.New(ctx, obj, event.FailOnNack(t))
	if err != nil {
		t.Fatalf("ttl test failed: %s\n", err.Error())
	}
	evts := f.Filter(e)
	if len(evts) != 1 {
		t.Fatalf("wrong number of ttled events: %d, now is %s\n", len(evts), tsStr)
	}
	expectedEventStr := eventStr
	var res interface{}
	err = json.Unmarshal([]byte(expectedEventStr), &res)
	if err != nil {
		t.Fatalf("ttl test failed: %s\n", err.Error())
	}
	if !reflect.DeepEqual(evts[0].Payload(), res) {
		pl, _ := json.MarshalIndent(evts[0].Payload(), "", "\t")
		t.Fatalf("wrong payload in ttled event: %s\n", pl)
	}
}
