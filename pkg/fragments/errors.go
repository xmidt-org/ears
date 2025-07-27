// SPDX-FileCopyrightText: 2025 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package fragments

import (
	"github.com/xmidt-org/ears/pkg/errs"
	"github.com/xmidt-org/ears/pkg/tenant"
)

func (e *InvalidFragmentError) Unwrap() error {
	return e.Err
}

func (e *InvalidFragmentError) Error() string {
	return errs.String("InvalidFragmentError", nil, e.Err)
}

type FragmentNotFoundError struct {
	TenantId     tenant.Id
	FragmentName string
}

func (e *FragmentNotFoundError) Error() string {
	return errs.String("FragmentNotFoundError", map[string]interface{}{"fragmentName": e.FragmentName, "orgId": e.TenantId.OrgId, "appId": e.TenantId.AppId}, nil)
}
