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
	"fmt"
	"github.com/xmidt-org/ears/pkg/event"
	"github.com/xmidt-org/ears/pkg/tenant"
	"sync"

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
