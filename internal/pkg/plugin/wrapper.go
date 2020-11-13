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

	"github.com/xmidt-org/ears/pkg/event"
	pkgfilter "github.com/xmidt-org/ears/pkg/filter"
	pkgmanager "github.com/xmidt-org/ears/pkg/plugin/manager"
	pkgreceiver "github.com/xmidt-org/ears/pkg/receiver"
	pkgsender "github.com/xmidt-org/ears/pkg/sender"
)

var _ pkgreceiver.Receiver = (*wrapper)(nil)
var _ pkgfilter.Filterer = (*wrapper)(nil)
var _ pkgsender.Sender = (*wrapper)(nil)

type wrapperType int

const (
	typeWrapperUnknown = iota
	typeWrapperSender
	typeWrapperReceiver
	typeWrapperFilterer
)

type wrapper struct {
	manager     *manager
	wrapperType wrapperType
	hash        string
	active      bool
	sender      pkgsender.Sender
	receiver    pkgreceiver.Receiver
	filterer    pkgfilter.Filterer
}

func (w *wrapper) Send(ctx context.Context, e event.Event) error {
	if w.wrapperType != typeWrapperSender || w.sender == nil {
		return &pkgmanager.NilPluginError{}
	}
	return w.sender.Send(ctx, e)
}

func (w *wrapper) Receive(ctx context.Context, next pkgreceiver.NextFn) error {
	if w.wrapperType != typeWrapperReceiver || w.receiver == nil {
		return &pkgmanager.NilPluginError{}
	}
	return w.receiver.Receive(ctx, next)
}

func (w *wrapper) Filter(ctx context.Context, e event.Event) ([]event.Event, error) {
	if w.wrapperType != typeWrapperFilterer || w.filterer == nil {
		return nil, &pkgmanager.NilPluginError{}
	}
	return w.filterer.Filter(ctx, e)
}

func (w *wrapper) Unregister() error {
	if w == nil || w.manager == nil {
		return &pkgmanager.NotRegisteredError{}
	}

	switch w.wrapperType {
	case typeWrapperReceiver:
		return w.manager.UnregisterReceiver(w)
	case typeWrapperFilterer:
		return w.manager.UnregisterFilter(w)
	case typeWrapperSender:
		return w.manager.UnregisterSender(w)
	}

	return &pkgmanager.NotRegisteredError{}
}
