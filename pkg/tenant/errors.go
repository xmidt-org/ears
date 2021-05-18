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
