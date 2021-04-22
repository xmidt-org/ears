package tenant

import (
	"context"
	"encoding/base64"
	"github.com/xmidt-org/ears/pkg/errs"
)

const delimiter = "."

type Id struct {
	OrgId string `json:"orgId,omitempty"`
	AppId string `json:"appId,omitempty"`
}

func (id Id) Equal(tid Id) bool {
	return id.OrgId == tid.OrgId && id.AppId == tid.AppId
}

//A string representation of tenant id that can be used has a key
//for certain data structure like a map
func (id Id) Key() string {
	return base64.StdEncoding.EncodeToString([]byte(id.OrgId)) + delimiter +
		base64.StdEncoding.EncodeToString([]byte(id.AppId))
}

//A string representation of tenant id + route id that
//can be used has a key for certain data structure like a map
func (id Id) KeyWithRoute(routeId string) string {
	return id.Key() + delimiter + base64.StdEncoding.EncodeToString([]byte(routeId))
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

type InvalidFormatError struct {
	format string
}

func (e *InvalidFormatError) Error() string {
	return errs.String("InvalidFormatError", map[string]interface{}{"value": e.format}, nil)
}
