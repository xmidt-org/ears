package tenant

import (
	"context"
)

type Id struct {
	OrgId string
	AppId string
}

func (id Id) String() string {
	return id.OrgId + "&" + id.AppId
}

type Config struct {
	Quota Quota `json:"quota"`
}

type Quota struct {
	EventsPerSec int `json:"eventsPerSec"`
}

type TenantStorer interface {
	GetAllOrgs(context.Context) ([]string, error)
	GetAllAppsInOrg(ctx context.Context, orgId string) ([]string, error)

	GetAppConfig(ctx context.Context, id Id) (Config, error)
	SetAppConfig(ctx context.Context, id Id, config Config) error
	DeleteAppConfig(ctx context.Context, id Id) error
}
