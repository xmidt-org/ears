package db_test

import (
	"context"
	"github.com/pkg/errors"
	"github.com/sebdah/goldie/v2"
	"github.com/xmidt-org/ears/pkg/tenant"
	"testing"
	"time"
)

type TestCase struct {
	config tenant.Config
}

var tenantTestCases = []TestCase{
	{
		config: tenant.Config{
			Tenant: tenant.Id{"myOrg", "myApp"},
			Quota: tenant.Quota{
				EventsPerSec: 10,
			},
		},
	},
	{
		config: tenant.Config{
			Tenant: tenant.Id{"myOrg", "myApp2"},
			Quota: tenant.Quota{
				EventsPerSec: 15,
			},
		},
	},
	{
		config: tenant.Config{
			Tenant: tenant.Id{"myOrg2", "myApp"},
			Quota: tenant.Quota{
				EventsPerSec: 30,
			},
		},
	},
}

func testTenantStorer(s tenant.TenantStorer, t *testing.T) {
	ctx := context.Background()

	//clear up the data first
	for _, tc := range tenantTestCases {
		s.DeleteConfig(ctx, tc.config.Tenant)
	}

	g := goldie.New(t)

	for _, tc := range tenantTestCases {
		var tenantNotFound *tenant.TenantNotFoundError
		_, err := s.GetConfig(ctx, tc.config.Tenant)
		if err != nil && !errors.As(err, &tenantNotFound) {
			t.Errorf("Expect tenantNotFound error, but get %s instead\n", err.Error())
		}

		err = s.SetConfig(ctx, tc.config)
		if err != nil {
			t.Errorf("Fail to set config %s\n", err.Error())
		}

		config, err := s.GetConfig(ctx, tc.config.Tenant)
		if err != nil {
			t.Errorf("Fail to get config %s\n", err.Error())
		}

		//confirm that modified time is recent within 10 seconds
		modified := time.Unix(config.Modified, 0)
		if modified.Add(10 * time.Second).Before(time.Now()) {
			t.Errorf("Motified time too old")
		}

		config.Modified = 0
		g.AssertJson(t, "config_"+tc.config.Tenant.OrgId+"_"+tc.config.Tenant.AppId, config)
	}

	for _, tc := range tenantTestCases {
		err := s.DeleteConfig(ctx, tc.config.Tenant)
		if err != nil {
			t.Errorf("Fail to delete app %s\n", err.Error())
		}

		var tenantNotFound *tenant.TenantNotFoundError
		_, err = s.GetConfig(ctx, tc.config.Tenant)
		if err != nil && !errors.As(err, &tenantNotFound) {
			t.Errorf("Expect tenantNotFound error, but get %s instead\n", err.Error())
		}
	}
}
