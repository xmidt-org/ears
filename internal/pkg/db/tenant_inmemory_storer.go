package db

import (
	"context"
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
	appsInOrg, ok := s.orgs[id.OrgId]
	if !ok {
		return nil, &tenant.TenantNotFoundError{id}
	}
	config, ok := appsInOrg[id.AppId]
	if !ok {
		return nil, &tenant.TenantNotFoundError{id}
	}
	return &config, nil
}

func (s *InMemoryStorer) SetConfig(ctx context.Context, config tenant.Config) error {
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
	appsInOrg, ok := s.orgs[id.OrgId]
	if !ok {
		return &tenant.TenantNotFoundError{id}
	}
	_, ok = appsInOrg[id.AppId]
	if !ok {
		return &tenant.TenantNotFoundError{id}
	}
	delete(appsInOrg, id.AppId)
	return nil
}
