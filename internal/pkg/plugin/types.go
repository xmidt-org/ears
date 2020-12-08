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
	"time"

	pkgfilter "github.com/xmidt-org/ears/pkg/filter"
	pkgmanager "github.com/xmidt-org/ears/pkg/plugin/manager"
	pkgreceiver "github.com/xmidt-org/ears/pkg/receiver"
	pkgsender "github.com/xmidt-org/ears/pkg/sender"
)

type Manager interface {
	Receiverers() map[string]pkgreceiver.NewReceiverer
	RegisterReceiver(
		ctx context.Context, plugin string,
		name string, config interface{},
	) (pkgreceiver.Receiver, error)
	Receivers() map[string]pkgreceiver.Receiver
	UnregisterReceiver(ctx context.Context, r pkgreceiver.Receiver) error

	Filterers() map[string]pkgfilter.NewFilterer
	RegisterFilter(
		ctx context.Context, plugin string,
		name string, config interface{},
	) (pkgfilter.Filterer, error)
	Filters() map[string]pkgfilter.Filterer
	UnregisterFilter(ctx context.Context, f pkgfilter.Filterer) error

	Senderers() map[string]pkgsender.NewSenderer
	RegisterSender(
		ctx context.Context, plugin string,
		name string, config interface{},
	) (pkgsender.Sender, error)
	Senders() map[string]pkgsender.Sender
	UnregisterSender(ctx context.Context, s pkgsender.Sender) error
}

type ManagerOption func(ManagerOptionProcessor) error

type ManagerOptionProcessor interface {
	WithPluginManager(p pkgmanager.Manager) error
	WithNextFnDeadline(d time.Duration) error
}

func WithPluginManager(p pkgmanager.Manager) ManagerOption {
	return func(o ManagerOptionProcessor) error {
		return o.WithPluginManager(p)
	}
}

func WithNextFnDeadline(d time.Duration) ManagerOption {
	return func(o ManagerOptionProcessor) error {
		return o.WithNextFnDeadline(d)
	}
}

type OptionError struct {
	Message string
	Err     error
}

type RegistrationError struct {
	Message string
	Plugin  string
	Name    string
	Err     error
}

type UnregistrationError struct {
	Message string
	Err     error
}

type NotRegisteredError struct{}
