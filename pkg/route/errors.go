// SPDX-FileCopyrightText: 2025 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package route

import (
	"github.com/xmidt-org/ears/pkg/errs"
	"github.com/xmidt-org/ears/pkg/tenant"
)

func (e *InvalidRouteError) Unwrap() error {
	return e.Err
}

func (e *InvalidRouteError) Error() string {
	return errs.String("InvalidRouteError", nil, e.Err)
}

type RouteNotFoundError struct {
	TenantId tenant.Id
	RouteId  string
}

func (e *RouteNotFoundError) Error() string {
	return errs.String("RouteNotFoundError", map[string]interface{}{"routeId": e.RouteId, "orgId": e.TenantId.OrgId, "appId": e.TenantId.AppId}, nil)
}
