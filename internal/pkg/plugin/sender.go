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
	pkgmanager "github.com/xmidt-org/ears/pkg/plugin/manager"
	pkgsender "github.com/xmidt-org/ears/pkg/sender"
)

var _ pkgsender.Sender = (*sender)(nil)

type sender struct {
	sync.Mutex

	id   string
	name string
	hash string

	manager *manager
	active  bool

	sender pkgsender.Sender
}

func (s *sender) Send(ctx context.Context, e event.Event) error {
	if s == nil {
		return &pkgmanager.NilPluginError{}
	}

	{
		s.Lock()
		if s.sender == nil || !s.active {
			s.Unlock()
			return &pkgmanager.NilPluginError{}
		}
		s.Unlock()
	}

	return s.sender.Send(ctx, e)
}

func (s *sender) Unregister(ctx context.Context) error {
	if s == nil {
		return &pkgmanager.NilPluginError{}
	}

	{
		s.Lock()
		if s.manager == nil || !s.active {
			s.Unlock()
			return &pkgmanager.NotRegisteredError{}
		}
		s.Unlock()
	}

	return s.manager.UnregisterSender(ctx, s)
}
