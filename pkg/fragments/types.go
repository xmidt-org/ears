// Copyright 2020 Comcast Cable Communications Management, LLC
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
