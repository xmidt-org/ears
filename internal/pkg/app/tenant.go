// SPDX-FileCopyrightText: 2025 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package app

type Tenant struct {
	orgId string
	appId string
}

func NewTenantId(orgId, appId string) *Tenant {
	return &Tenant{
		orgId: orgId,
		appId: appId,
	}
}

func (t *Tenant) OrgId() string {
	return t.orgId
}

func (t *Tenant) AppId() string {
	return t.appId
}
