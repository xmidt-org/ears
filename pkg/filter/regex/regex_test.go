// SPDX-FileCopyrightText: 2025 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package regex_test

import (
	"context"
	"encoding/json"
	"reflect"
	"testing"

	"github.com/xmidt-org/ears/pkg/event"
	"github.com/xmidt-org/ears/pkg/filter/regex"
	"github.com/xmidt-org/ears/pkg/tenant"
)

func TestFilterHashBasic(t *testing.T) {
	ctx := context.Background()
	f, err := regex.NewFilter(tenant.Id{AppId: "myapp", OrgId: "myorg"}, "regex", "myregex", regex.Config{
		FromPath: ".foo",
		ToPath:   ".regex",
		Regex:    "[\\w]+$",
	}, nil, nil)
	if err != nil {
		t.Fatalf("regex test failed: %s\n", err.Error())
	}
	eventStr := `{ "foo": "foo-bar"}`
	var obj interface{}
	err = json.Unmarshal([]byte(eventStr), &obj)
	if err != nil {
		t.Fatalf("regex test failed: %s\n", err.Error())
	}
	e, err := event.New(ctx, obj, event.FailOnNack(t))
	if err != nil {
		t.Fatalf("regex test failed: %s\n", err.Error())
	}
	evts := f.Filter(e)
	if len(evts) != 1 {
		t.Fatalf("wrong number of regexhashed events: %d\n", len(evts))
	}
	expectedEventStr := `{"foo":"foo-bar", "regex":"bar"}`
	var res interface{}
	err = json.Unmarshal([]byte(expectedEventStr), &res)
	if err != nil {
		t.Fatalf("regex test failed: %s\n", err.Error())
	}
	if !reflect.DeepEqual(evts[0].Payload(), res) {
		pl, _ := json.MarshalIndent(evts[0].Payload(), "", "\t")
		t.Fatalf("wrong payload in regexhashed event: %s\n", pl)
	}
}
