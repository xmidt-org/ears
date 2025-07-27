// SPDX-FileCopyrightText: 2025 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package fragments

import (
	"context"

	"github.com/xmidt-org/ears/pkg/route"
	"github.com/xmidt-org/ears/pkg/tenant"
)

type InvalidFragmentError struct {
	Err error
}

// FragmentStorer stores fragments of route configuration. A fragment is a plugin configuration
// identified by its unique name and tenant. A fragment can represent a receiver, sender or filter
// plugin configuration and is always of type route.PluginConfig.
type FragmentStorer interface {
	GetAllFragments(context.Context) ([]route.PluginConfig, error)
	GetFragment(context.Context, tenant.Id, string) (route.PluginConfig, error)
	GetAllTenantFragments(ctx context.Context, id tenant.Id) ([]route.PluginConfig, error)
	SetFragment(context.Context, tenant.Id, route.PluginConfig) error
	SetFragments(context.Context, tenant.Id, []route.PluginConfig) error
	DeleteFragment(context.Context, tenant.Id, string) error
	DeleteFragments(context.Context, tenant.Id, []string) error
}
