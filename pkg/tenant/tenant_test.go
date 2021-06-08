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

package tenant_test

import (
	"github.com/sebdah/goldie/v2"
	"github.com/xmidt-org/ears/pkg/tenant"
	"testing"
)

func TestTenantId(t *testing.T) {
	id1 := tenant.Id{OrgId: "myOrg", AppId: "myApp"}
	id2 := tenant.Id{OrgId: "myOrg", AppId: "myApp"}
	id3 := tenant.Id{OrgId: "myOrg", AppId: "myApp2"}

	if !id1.Equal(id2) {
		t.Errorf("Expect id1 and id2 are equal")
	}
	if id1.Equal(id3) {
		t.Errorf("Expect id1 and id3 are not equal")
	}

	g := goldie.New(t)

	g.Assert(t, "key", []byte(id1.Key()))
	g.Assert(t, "string", []byte(id1.ToString()))
	g.Assert(t, "keyRoute", []byte(id1.KeyWithRoute("routeId")))
}
