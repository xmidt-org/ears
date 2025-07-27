// SPDX-FileCopyrightText: 2025 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package plugin

import (
	"context"
	"sync"

	"github.com/xmidt-org/ears/pkg/tenant"

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

func (s *sender) EventTs() int64 {
	return s.sender.EventTs()
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
