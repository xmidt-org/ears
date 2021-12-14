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

package ttl_test

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/xmidt-org/ears/pkg/event"
	"github.com/xmidt-org/ears/pkg/filter/ttl"
	"github.com/xmidt-org/ears/pkg/tenant"
	"github.com/xorcare/pointer"
	"reflect"
	"testing"
	"time"
)

func TestFilterTtlBasic(t *testing.T) {
	ctx := context.Background()
	f, err := ttl.NewFilter(tenant.Id{AppId: "myapp", OrgId: "myorg"}, "ttl", "myttl", ttl.Config{
		Path:       ".ts",
		NanoFactor: pointer.Int(1),
		Ttl:        pointer.Int(300000),
	}, nil)
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
		t.Fatalf("wrong number of ttled events: %d\n", len(evts))
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
