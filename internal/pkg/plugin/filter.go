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

package plugin

import (
	"context"
	"github.com/xmidt-org/ears/pkg/tenant"
	"sync"

	"github.com/xmidt-org/ears/pkg/event"
	pkgfilter "github.com/xmidt-org/ears/pkg/filter"
	pkgmanager "github.com/xmidt-org/ears/pkg/plugin/manager"
)

var _ pkgfilter.Filterer = (*filter)(nil)

type filter struct {
	sync.Mutex

	id     string
	name   string
	plugin string
	hash   string
	tid    tenant.Id

	manager *manager

	active   bool
	filterer pkgfilter.Filterer
}

func (f *filter) Filter(e event.Event) []event.Event {
	if f.filterer == nil {
		e.Nack(&pkgmanager.NilPluginError{})
		return nil
	}
	{
		f.Lock()
		if !f.active {
			f.Unlock()
			e.Nack(&pkgmanager.NotRegisteredError{})
			return nil
		}
		f.Unlock()
	}
	return f.filterer.Filter(e)
}

func (f *filter) Unregister(ctx context.Context) error {
	if f == nil {
		return &pkgmanager.NilPluginError{}
	}
	{
		f.Lock()
		if f.manager == nil || !f.active {
			f.Unlock()
			return &pkgmanager.NotRegisteredError{}
		}
		f.Unlock()
	}
	return f.manager.UnregisterFilter(ctx, f)
}

func (f *filter) Config() interface{} {
	return f.filterer.Config()
}

func (f *filter) Name() string {
	return f.name
}

func (f *filter) Plugin() string {
	return f.plugin
}

func (f *filter) Tenant() tenant.Id {
	return f.tid
}

func (f *filter) EventSuccessCount() int {
	return f.EventSuccessCount()
}

func (f *filter) EventSuccessVelocity() int {
	return f.EventSuccessVelocity()
}

func (f *filter) EventFilterCount() int {
	return f.EventFilterCount()
}

func (f *filter) EventFilterVelocity() int {
	return f.EventFilterVelocity()
}

func (f *filter) EventErrorCount() int {
	return f.EventErrorCount()
}

func (f *filter) EventErrorVelocity() int {
	return f.EventErrorVelocity()
}

func (f *filter) EventTs() int64 {
	return f.EventTs()
}
