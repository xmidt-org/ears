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
