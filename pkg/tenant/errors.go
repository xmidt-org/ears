package tenant

import "github.com/xmidt-org/ears/pkg/errs"

type TenantNotFoundError struct {
	Tenant Id
}

func (e *TenantNotFoundError) Error() string {
	return errs.String("TenantNotFoundError", map[string]interface{}{"orgId": e.Tenant.OrgId, "appId": e.Tenant.AppId}, nil)
}
