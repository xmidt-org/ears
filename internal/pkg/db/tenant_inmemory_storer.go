// SPDX-FileCopyrightText: 2025 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package db

import (
	"context"
	"time"

	"github.com/xmidt-org/ears/internal/pkg/rtsemconv"
	"github.com/xmidt-org/ears/pkg/tenant"
)

type InMemoryStorer struct {
	orgs map[string]map[string]tenant.Config
}

func NewTenantInmemoryStorer() *InMemoryStorer {
	return &InMemoryStorer{
		make(map[string]map[string]tenant.Config),
	}
}

func (s *InMemoryStorer) GetAllConfigs(ctx context.Context) ([]tenant.Config, error) {
	_, span := CreateSpan(ctx, "getAllTenantConfigs", rtsemconv.DBSystemInMemory)
	defer span.End()
	configs := make([]tenant.Config, 0)
	for _, org := range s.orgs {
		for _, appConfig := range org {
			configs = append(configs, appConfig)
		}
	}
	return configs, nil
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
