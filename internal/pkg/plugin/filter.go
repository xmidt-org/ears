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
	"sync"

	"github.com/xmidt-org/ears/pkg/event"
	pkgfilter "github.com/xmidt-org/ears/pkg/filter"
	pkgmanager "github.com/xmidt-org/ears/pkg/plugin/manager"
)

var _ pkgfilter.Filterer = (*filter)(nil)

type filter struct {
	sync.Mutex

	id   string
	name string
	hash string

	manager *manager

	active   bool
	filterer pkgfilter.Filterer
}

func (f *filter) Filter(ctx context.Context, e event.Event) ([]event.Event, error) {
	if f.filterer == nil {
		return nil, &pkgmanager.NilPluginError{}
	}

	{
		f.Lock()
		if !f.active {
			f.Unlock()
			return nil, &pkgmanager.NotRegisteredError{}
		}
		f.Unlock()
	}

	return f.filterer.Filter(ctx, e)
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
