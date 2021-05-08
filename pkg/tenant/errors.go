package tenant

import "github.com/xmidt-org/ears/pkg/errs"

// Standard set of errors for TenantStorer

type TenantNotFoundError struct {
	Tenant Id
}

func (e *TenantNotFoundError) Error() string {
	return errs.String("TenantNotFoundError", map[string]interface{}{"orgId": e.Tenant.OrgId, "appId": e.Tenant.AppId}, nil)
}

type BadConfigError struct {
}

func (e *BadConfigError) Error() string {
	return errs.String("BadConfigError", nil, nil)
}

type InternalStorageError struct {
	Wrapped error
}

func (e *InternalStorageError) Error() string {
	return errs.String("InternalStorageError", nil, e.Wrapped)
}

func (e *InternalStorageError) Unwrap() error {
	return e.Wrapped
}
