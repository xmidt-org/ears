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

package db

import (
	"context"
	"github.com/xmidt-org/ears/internal/pkg/rtsemconv"
	"github.com/xmidt-org/ears/pkg/tenant"
	"time"
)

type InMemoryStorer struct {
	orgs map[string]map[string]tenant.Config
}

func NewTenantInmemoryStorer() *InMemoryStorer {
	return &InMemoryStorer{
		make(map[string]map[string]tenant.Config),
	}
}

func (s *InMemoryStorer) GetConfig(ctx context.Context, id tenant.Id) (*tenant.Config, error) {

	_, span := CreateSpan(ctx, "getTenantConfig", rtsemconv.DBSystemInMemory)
	defer span.End()

	appsInOrg, ok := s.orgs[id.OrgId]
	if !ok {
		return nil, &tenant.TenantNotFoundError{Tenant: id}
	}
	config, ok := appsInOrg[id.AppId]
	if !ok {
		return nil, &tenant.TenantNotFoundError{Tenant: id}
	}
	return &config, nil
}

func (s *InMemoryStorer) SetConfig(ctx context.Context, config tenant.Config) error {

	_, span := CreateSpan(ctx, "setTenantConfig", rtsemconv.DBSystemInMemory)
	defer span.End()

	appsInOrg, ok := s.orgs[config.Tenant.OrgId]
	if !ok {
		appsInOrg = make(map[string]tenant.Config)
		s.orgs[config.Tenant.OrgId] = appsInOrg
	}
	config.Modified = time.Now().Unix()
	appsInOrg[config.Tenant.AppId] = config
	return nil
}

func (s *InMemoryStorer) DeleteConfig(ctx context.Context, id tenant.Id) error {

	_, span := CreateSpan(ctx, "deleteTenantConfig", rtsemconv.DBSystemInMemory)
	defer span.End()

	appsInOrg, ok := s.orgs[id.OrgId]
	if !ok {
		return &tenant.TenantNotFoundError{Tenant: id}
	}
	_, ok = appsInOrg[id.AppId]
	if !ok {
		return &tenant.TenantNotFoundError{Tenant: id}
	}
	delete(appsInOrg, id.AppId)
	return nil
}
