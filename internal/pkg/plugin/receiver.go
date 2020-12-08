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
	"sync"

	pkgreceiver "github.com/xmidt-org/ears/pkg/receiver"
)

var _ pkgreceiver.Receiver = (*receiver)(nil)

type wrapperType int

type receiver struct {
	sync.Mutex

	id   string
	name string
	hash string

	manager *manager
	active  bool

	receiver pkgreceiver.Receiver

	done chan struct{}
}

func (r *receiver) Receive(ctx context.Context, next pkgreceiver.NextFn) error {
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

	// Block
	return r.manager.receive(ctx, r, next)
}

func (r *receiver) StopReceiving(ctx context.Context) error {

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
	{
		r.Lock()
		if r == nil || r.manager == nil || !r.active {
			r.Unlock()
			return &NotRegisteredError{}
		}
		r.Unlock()
	}

	return r.manager.UnregisterReceiver(ctx, r)
}
