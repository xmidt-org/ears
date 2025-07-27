// SPDX-FileCopyrightText: 2025 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package hash_test

import (
	"context"
	"encoding/json"
	"reflect"
	"testing"

	"github.com/xmidt-org/ears/pkg/event"
	"github.com/xmidt-org/ears/pkg/filter/hash"
	"github.com/xmidt-org/ears/pkg/tenant"
)

func TestFilterHashBasic(t *testing.T) {
	ctx := context.Background()
	f, err := hash.NewFilter(tenant.Id{AppId: "myapp", OrgId: "myorg"}, "hash", "myhash", hash.Config{
		FromPath:      ".foo",
		ToPath:        ".hash",
		HashAlgorithm: "sha1",
		Encoding:      "hex",
	}, nil, nil)
	if err != nil {
		t.Fatalf("hash test failed: %s\n", err.Error())
	}
	eventStr := `{ "foo": "bar"}`
	var obj interface{}
	err = json.Unmarshal([]byte(eventStr), &obj)
	if err != nil {
		t.Fatalf("hash test failed: %s\n", err.Error())
	}
	e, err := event.New(ctx, obj, event.FailOnNack(t))
	if err != nil {
		t.Fatalf("hash test failed: %s\n", err.Error())
	}
	evts := f.Filter(e)
	if len(evts) != 1 {
		t.Fatalf("wrong number of hashed events: %d\n", len(evts))
	}
	expectedEventStr := `{"foo":"bar", "hash":"62cdb7020ff920e5aa642c3d4066950dd1f01f4d"}`
	var res interface{}
	err = json.Unmarshal([]byte(expectedEventStr), &res)
	if err != nil {
		t.Fatalf("hash test failed: %s\n", err.Error())
	}
	if !reflect.DeepEqual(evts[0].Payload(), res) {
		pl, _ := json.MarshalIndent(evts[0].Payload(), "", "\t")
		t.Fatalf("wrong payload in hashed event: %s\n", pl)
	}
}
