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
	pkgmanager "github.com/xmidt-org/ears/pkg/plugin/manager"
	pkgsender "github.com/xmidt-org/ears/pkg/sender"
)

var _ pkgsender.Sender = (*sender)(nil)

type sender struct {
	sync.Mutex
	id      string
	tid     tenant.Id
	name    string
	plugin  string
	hash    string
	manager *manager
	active  bool
	sender  pkgsender.Sender
}

func (s *sender) Unwrap() pkgsender.Sender {
	return s.sender
}

func (s *sender) Config() interface{} {
	return s.sender.Config()
}

func (s *sender) Name() string {
	return s.name
}

func (s *sender) Plugin() string {
	return s.plugin
}

func (s *sender) Tenant() tenant.Id {
	return s.tid
}

func (s *sender) EventSuccessCount() int {
	return s.sender.EventSuccessCount()
}

func (s *sender) EventSuccessVelocity() int {
	return s.sender.EventSuccessVelocity()
}

func (s *sender) EventErrorCount() int {
	return s.sender.EventErrorCount()
}

func (s *sender) EventErrorVelocity() int {
	return s.sender.EventErrorVelocity()
}

func (s *sender) Send(e event.Event) {
	if s == nil {
		e.Nack(&pkgmanager.NilPluginError{})
		return
	}
	{
		s.Lock()
		if s.sender == nil || !s.active {
			s.Unlock()
			e.Nack(&pkgmanager.NilPluginError{})
			return
		}
		s.Unlock()
	}
	s.sender.Send(e)
}

func (s *sender) StopSending(ctx context.Context) {
	s.sender.StopSending(ctx)
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
