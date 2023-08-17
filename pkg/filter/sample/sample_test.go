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

package sample_test

import (
	"context"
	"encoding/json"
	"github.com/xmidt-org/ears/pkg/event"
	"github.com/xmidt-org/ears/pkg/filter/sample"
	"github.com/xmidt-org/ears/pkg/tenant"
	"github.com/xorcare/pointer"
	"testing"
)

func TestFilterDecodeBasic(t *testing.T) {
	ctx := context.Background()
	f, err := sample.NewFilter(tenant.Id{AppId: "myapp", OrgId: "myorg"}, "sample", "mysample", sample.Config{
		Percentage: pointer.Float64(0.1),
	}, nil, nil)
	if err != nil {
		t.Fatalf("sample test failed: %s\n", err.Error())
	}
	eventStr := `{ "foo": "bar"}`
	var obj interface{}
	err = json.Unmarshal([]byte(eventStr), &obj)
	if err != nil {
		t.Fatalf("sample test failed: %s\n", err.Error())
	}
	e, err := event.New(ctx, obj, event.FailOnNack(t))
	if err != nil {
		t.Fatalf("sample test failed: %s\n", err.Error())
	}
	evtCnt := 0
	for i := 0; i < 1000; i++ {
		evts := f.Filter(e)
		evtCnt += len(evts)
	}
	if evtCnt < 80 || evtCnt > 120 {
		t.Fatalf("unexpected number of sampled events: %d\n", evtCnt)
	}
}
