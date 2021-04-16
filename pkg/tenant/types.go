package tenant

import (
	"context"
	"github.com/xmidt-org/ears/pkg/errs"
	"strings"
)

const delimiter = "&"

type Id struct {
	OrgId string `json:"orgId,omitempty"`
	AppId string `json:"appId,omitempty"`
}

func FromString(str string) (*Id, error) {
	elems := strings.Split(str, delimiter)
	if len(elems) != 2 {
		return nil, &InvalidFormatError{str}
	}
	return &Id{elems[0], elems[1]}, nil
}

func (id Id) String() string {
	return id.OrgId + delimiter + id.AppId
}

func (id Id) Equal(tid Id) bool {
	return id.OrgId == tid.OrgId && id.AppId == tid.AppId
}

//A string representation of tenant id + route id
//Can be used has a key for certain data structure like a map
func (id Id) Key(routeId string) string {
	return id.String() + delimiter + routeId
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
