// SPDX-FileCopyrightText: 2025 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package plugin

import (
	"context"
	"fmt"
	"sync"

	"github.com/xmidt-org/ears/pkg/event"
	"github.com/xmidt-org/ears/pkg/tenant"

	pkgmanager "github.com/xmidt-org/ears/pkg/plugin/manager"
	pkgreceiver "github.com/xmidt-org/ears/pkg/receiver"
)

var _ pkgreceiver.Receiver = (*receiver)(nil)

type receiver struct {
	sync.Mutex

	id     string
	name   string
	plugin string
	hash   string
	tid    tenant.Id
	next   pkgreceiver.NextFn

	manager *manager
	active  bool

	receiver pkgreceiver.Receiver

	done chan struct{}
}

func (r *receiver) Config() interface{} {
	return r.receiver.Config()
}

func (r *receiver) Name() string {
	return r.name
}

func (r *receiver) Plugin() string {
	return r.plugin
}

func (r *receiver) Tenant() tenant.Id {
	return r.tid
}

func (r *receiver) EventSuccessCount() int {
	return r.receiver.EventSuccessCount()
}

func (r *receiver) EventSuccessVelocity() int {
	return r.receiver.EventSuccessVelocity()
}

func (r *receiver) EventErrorCount() int {
	return r.receiver.EventErrorCount()
}

func (r *receiver) EventErrorVelocity() int {
	return r.receiver.EventErrorVelocity()
}

func (r *receiver) EventTs() int64 {
	return r.receiver.EventTs()
}

func (r *receiver) LogSuccess() {
	r.receiver.LogSuccess()
}

func (r *receiver) Trigger(e event.Event) {
	//r.receiver.Trigger(e)
	r.Lock()
	next := r.next
	r.Unlock()
	if next != nil {
		next(e)
	}
}

func (r *receiver) Receive(next pkgreceiver.NextFn) error {
	if r == nil {
		return &pkgmanager.NilPluginError{}
	}

	if next == nil {
		return &pkgreceiver.InvalidConfigError{
			Err: fmt.Errorf("next cannot be nil"),
		}
	}

	{
		r.Lock()
		if !r.active {
			r.Unlock()
			return &NotRegisteredError{}
		}
		r.Unlock()
	}

	r.next = next

	// Block
	return r.manager.receive(r, next)
}

func (r *receiver) StopReceiving(ctx context.Context) error {
	if r == nil {
		return &pkgmanager.NilPluginError{}
	}

	{
		r.Lock()
		if !r.active {
			r.Unlock()
			return &NotRegisteredError{}
		}
		r.Unlock()
	}

	return r.manager.stopReceiving(ctx, r)
}

func (r *receiver) Unregister(ctx context.Context) error {
	if r == nil {
		return &pkgmanager.NilPluginError{}
	}

	{
		r.Lock()
		if r.manager == nil || !r.active {
			r.Unlock()
			return &NotRegisteredError{}
		}
		r.Unlock()
	}

	return r.manager.UnregisterReceiver(ctx, r)
}
