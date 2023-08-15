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
	"github.com/rs/zerolog"
	"github.com/xmidt-org/ears/internal/pkg/quota"
	"github.com/xmidt-org/ears/pkg/secret"
	"github.com/xmidt-org/ears/pkg/tenant"
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
		tid tenant.Id,
	) (pkgreceiver.Receiver, error)
	Receivers() map[string]pkgreceiver.Receiver
	ReceiversStatus() map[string]ReceiverStatus
	UnregisterReceiver(ctx context.Context, r pkgreceiver.Receiver) error

	Filterers() map[string]pkgfilter.NewFilterer
	RegisterFilter(
		ctx context.Context, plugin string,
		name string, config interface{},
		tid tenant.Id,
	) (pkgfilter.Filterer, error)
	Filters() map[string]pkgfilter.Filterer
	FiltersStatus() map[string]FilterStatus
	UnregisterFilter(ctx context.Context, f pkgfilter.Filterer) error

	Senderers() map[string]pkgsender.NewSenderer
	RegisterSender(
		ctx context.Context, plugin string,
		name string, config interface{},
		tid tenant.Id,
	) (pkgsender.Sender, error)
	Senders() map[string]pkgsender.Sender
	SendersStatus() map[string]SenderStatus
	UnregisterSender(ctx context.Context, s pkgsender.Sender) error
}

type ManagerOption func(*manager) error

func WithPluginManager(p pkgmanager.Manager) ManagerOption {
	return func(m *manager) error {
		if p == nil {
			return &OptionError{
				Message: "plugin manager cannot be nil",
			}
		}
		m.pm = p
		return nil
	}
}

func WithLogger(l *zerolog.Logger) ManagerOption {
	return func(m *manager) error {
		m.logger = l
		return nil
	}
}

func WithTenantStorer(t tenant.TenantStorer) ManagerOption {
	return func(m *manager) error {
		m.tenantStorer = t
		return nil
	}
}

func WithQuotaManager(q *quota.QuotaManager) ManagerOption {
	return func(m *manager) error {
		m.quotaManager = q
		return nil
	}
}

func WithSecretVaults(s secret.Vault) ManagerOption {
	return func(m *manager) error {
		m.secrets = s
		return nil
	}
}

func WithNextFnDeadline(d time.Duration) ManagerOption {
	return func(m *manager) error {
		m.nextFnDeadline = d
		return nil
	}
}

type ReceiverStatus struct {
	Name            string
	Plugin          string
	Config          interface{}
	ReferenceCount  int
	SuccessCount    int
	ErrorCount      int
	SuccessVelocity int
	ErrorVelocity   int
	LastEventTs     int64
	Tid             tenant.Id
}

type SenderStatus struct {
	Name            string
	Plugin          string
	Config          interface{}
	ReferenceCount  int
	SuccessCount    int
	ErrorCount      int
	SuccessVelocity int
	ErrorVelocity   int
	LastEventTs     int64
	Tid             tenant.Id
}

type FilterStatus struct {
	Name           string
	Plugin         string
	Config         interface{}
	ReferenceCount int
	Tid            tenant.Id
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
