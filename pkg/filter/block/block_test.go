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

package block_test

import (
	"context"
	"encoding/json"
	"github.com/xmidt-org/ears/pkg/event"
	"github.com/xmidt-org/ears/pkg/filter/block"
	"github.com/xmidt-org/ears/pkg/tenant"
	"testing"
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
