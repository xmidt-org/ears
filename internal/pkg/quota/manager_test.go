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

package quota_test

import (
	"context"
	"errors"
	"github.com/rs/zerolog"
	"github.com/spf13/viper"
	"github.com/xmidt-org/ears/internal/pkg/config"
	"github.com/xmidt-org/ears/internal/pkg/db"
	"github.com/xmidt-org/ears/internal/pkg/quota"
	"github.com/xmidt-org/ears/internal/pkg/syncer"
	"github.com/xmidt-org/ears/pkg/tenant"
	"os"
	"testing"
	"time"
)

func testConfig() config.Config {
	v := viper.New()
	v.Set("ears.logLevel", "info")
	v.Set("ears.api.port", 8080)
	v.Set("ears.storage.route.type", "inmemory")
	v.Set("ears.storage.tenant.type", "inmemory")
	v.Set("ears.synchronization.type", "inmemory")
	v.Set("ears.synchronization.active", "true")
	v.Set("ears.ratelimiter.type", "inmemory")
	return v
}

func setup(tenantStorer tenant.TenantStorer) (*quota.QuotaManager, error) {
	testLogger := zerolog.New(os.Stdout)

	config := testConfig()
	syncer := syncer.NewInMemoryDeltaSyncer(&testLogger, config)
	syncer.StartListeningForSyncRequests()
	return quota.NewQuotaManager(&testLogger, tenantStorer, syncer, config)
}

func TestQuotaManagerSyncing(t *testing.T) {

	//setup tenant shared tenant storer
	tenantStorer := db.NewTenantInmemoryStorer()
	ctx := context.Background()
	tenantConfig := tenant.Config{
		Tenant: tenant.Id{
			OrgId: "myOrg",
			AppId: "myApp",
		},
		Quota: tenant.Quota{
			EventsPerSec: 10,
		},
	}
	tenantStorer.SetConfig(ctx, tenantConfig)

	quotaMgr1, err := setup(tenantStorer)
	if err != nil {
		t.Fatalf("Fail to start quota manager 1 %s\n", err.Error())
	}
	quotaMgr2, err := setup(tenantStorer)
	if err != nil {
		t.Fatalf("Fail to start quota manager 2 %s\n", err.Error())
	}

	quotaMgr1.Start()
	quotaMgr2.Start()

	quotaMgr1.SyncItem(ctx, tenantConfig.Tenant, "", true)
	if quotaMgr1.TenantLimit(ctx, tenantConfig.Tenant) != 10 {
		t.Fatalf("Expect limit=10 instead, got limit=%d\n", quotaMgr1.TenantLimit(ctx, tenantConfig.Tenant))
	}
	if quotaMgr2.TenantLimit(ctx, tenantConfig.Tenant) != 10 {
		t.Fatalf("Expect limit=10 instead, got limit=%d\n", quotaMgr2.TenantLimit(ctx, tenantConfig.Tenant))
	}

	//update the limit to 20
	tenantConfig = tenant.Config{
		Tenant: tenant.Id{
			OrgId: "myOrg",
			AppId: "myApp",
		},
		Quota: tenant.Quota{
			EventsPerSec: 20,
		},
	}
	tenantStorer.SetConfig(ctx, tenantConfig)

	err = quotaMgr1.PublishQuota(ctx, tenantConfig.Tenant)
	if err != nil {
		t.Fatalf("Error publishing quota %s\n", err.Error())
	}
	time.Sleep(time.Second)
	if quotaMgr1.TenantLimit(ctx, tenantConfig.Tenant) != 20 {
		t.Fatalf("Expect limit=20 instead, got limit=%d\n", quotaMgr1.TenantLimit(ctx, tenantConfig.Tenant))
	}
	if quotaMgr2.TenantLimit(ctx, tenantConfig.Tenant) != 20 {
		t.Fatalf("Expect limit=20 instead, got limit=%d\n", quotaMgr2.TenantLimit(ctx, tenantConfig.Tenant))
	}

	quotaMgr1.Stop()
	quotaMgr2.Stop()
}

func TestQuotaManagerRateLimit(t *testing.T) {

	//setup tenant storer
	tenantStorer := db.NewTenantInmemoryStorer()
	ctx := context.Background()
	tenantConfig := tenant.Config{
		Tenant: tenant.Id{
			OrgId: "myOrg",
			AppId: "myApp",
		},
		Quota: tenant.Quota{
			EventsPerSec: 10,
		},
	}
	tenantStorer.SetConfig(ctx, tenantConfig)

	//setup quota manager
	quotaMgr, err := setup(tenantStorer)
	if err != nil {
		t.Fatalf("Fail to start quota manager %s\n", err.Error())
	}
	quotaMgr.Start()

	quotaMgr.SyncItem(ctx, tenantConfig.Tenant, "", true)

	for i := 0; i < 3; i++ {
		//validate that we can do 8rps
		err = validateQuotaMgrRps(quotaMgr, tenantConfig.Tenant, 8)
		if err == nil {
			break
		}
		if errors.Is(err, TestErr_FailToReachRps) {
			continue
		}
		t.Fatalf("Fail to reach the desired rps, error=%s\n", err.Error())
	}
	if err != nil {
		t.Fatalf("Fail to reach the desired rps in 3 tries")
	}

	quotaMgr.Stop()
}

var TestErr_FailToReachRps = errors.New("Cannot reach desired RPS")

func validateQuotaMgrRps(mgr *quota.QuotaManager, tid tenant.Id, rps int) error {
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	start := time.Now()

	count := 0
	for i := 0; i < rps; i++ {
		err := mgr.Wait(ctx, tid)
		if err != nil {
			return err
		}
		count++
	}

	if start.Add(time.Second + time.Millisecond*100).After(time.Now()) {
		return nil
	}
	return TestErr_FailToReachRps
}
