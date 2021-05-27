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
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/xmidt-org/ears/internal/pkg/logs"
	"github.com/xmidt-org/ears/internal/pkg/quota"
	"github.com/xmidt-org/ears/pkg/panics"
	"github.com/xmidt-org/ears/pkg/tenant"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/xmidt-org/ears/pkg/event"
	pkgevent "github.com/xmidt-org/ears/pkg/event"
	pkgfilter "github.com/xmidt-org/ears/pkg/filter"
	pkgmanager "github.com/xmidt-org/ears/pkg/plugin/manager"
	pkgreceiver "github.com/xmidt-org/ears/pkg/receiver"
	pkgsender "github.com/xmidt-org/ears/pkg/sender"
)

type manager struct {
	sync.Mutex

	pm pkgmanager.Manager

	receivers        map[string]pkgreceiver.Receiver
	receiversCount   map[string]int
	receiversWrapped map[string]*receiver
	receiversFn      map[string]map[string]pkgreceiver.NextFn // map[receiverKey]map[wrapperID] -> nextFN

	filters        map[string]pkgfilter.Filterer
	filtersCount   map[string]int
	filtersWrapped map[string]*filter

	senders        map[string]pkgsender.Sender
	sendersCount   map[string]int
	sendersWrapped map[string]*sender

	nextFnDeadline time.Duration

	logger *zerolog.Logger

	quotaManager *quota.QuotaManager
}

// === Initialization ================================================

const (
	defaultNextFnDeadline = 5 * time.Second
)

func NewManager(options ...ManagerOption) (Manager, error) {
	m := manager{
		receivers:        map[string]pkgreceiver.Receiver{},
		receiversCount:   map[string]int{},
		receiversWrapped: map[string]*receiver{},
		receiversFn:      map[string]map[string]pkgreceiver.NextFn{},
		nextFnDeadline:   defaultNextFnDeadline,

		filters:        map[string]pkgfilter.Filterer{},
		filtersCount:   map[string]int{},
		filtersWrapped: map[string]*filter{},

		senders:        map[string]pkgsender.Sender{},
		sendersCount:   map[string]int{},
		sendersWrapped: map[string]*sender{},
	}

	var err error
	for _, option := range options {
		err = option(&m)
		if err != nil {
			return nil, &OptionError{Err: err}
		}
	}

	return &m, nil
}

// === Receivers =====================================================

func (m *manager) Receiverers() map[string]pkgreceiver.NewReceiverer {
	if m.pm == nil {
		return map[string]pkgreceiver.NewReceiverer{}
	}
	return m.pm.Receiverers()
}

func (m *manager) RegisterReceiver(
	ctx context.Context, plugin string,
	name string, config interface{},
	tid tenant.Id,
) (pkgreceiver.Receiver, error) {

	ns, err := m.pm.Receiverer(plugin)
	if err != nil {
		return nil, &RegistrationError{
			Message: "could not get plugin",
			Plugin:  plugin,
			Name:    name,
			Err:     err,
		}
	}

	hash, err := ns.ReceiverHash(config)
	if err != nil {
		return nil, &RegistrationError{
			Message: "could not generate hash",
			Plugin:  plugin,
			Name:    name,
			Err:     err,
		}
	}

	key := m.mapkey(name, hash)

	m.Lock()
	defer m.Unlock()

	r, ok := m.receivers[key]
	if !ok {
		r, err = ns.NewReceiver(config)
		if err != nil {
			return nil, &RegistrationError{
				Message: "could not create new receiver",
				Plugin:  plugin,
				Name:    name,
				Err:     err,
			}
		}

		m.receivers[key] = r
		m.receiversCount[key] = 0
		m.receiversFn[key] = map[string]pkgreceiver.NextFn{}

		go func() {
			r.Receive(func(e event.Event) {
				if m.quotaManager != nil {
					//ratelimit
					err = m.quotaManager.Wait(e.Context(), tid)
					if err != nil {
						e.Nack(err)
						return
					}
				}
				m.next(key, e)
			})
		}()
	}

	u, err := uuid.NewRandom()
	if err != nil {
		return nil, &RegistrationError{
			Message: "could not generate unique id",
			Plugin:  plugin,
			Name:    name,
			Err:     err,
		}
	}

	w := &receiver{
		id:       u.String(),
		name:     name,
		plugin:   plugin,
		hash:     hash,
		manager:  m,
		receiver: r,
		active:   true,
	}

	m.receiversWrapped[w.id] = w
	m.receiversCount[key]++

	return w, nil

}

func (m *manager) Receivers() map[string]pkgreceiver.Receiver {
	m.Lock()
	defer m.Unlock()
	receivers := map[string]pkgreceiver.Receiver{}
	for k, v := range m.receiversWrapped {
		receivers[k] = v
	}
	return receivers
}

func (m *manager) UniqueReceivers() map[string]pkgreceiver.Receiver {
	m.Lock()
	defer m.Unlock()
	receivers := map[string]pkgreceiver.Receiver{}
	for _, v := range m.receiversWrapped {
		receivers[v.hash] = v
	}
	return receivers
}

// next iterates through all receiver functions that have registered for
// a receiver (unique by name + config hash).  These must be independent,
// so no error can actually be returned to the receiver if a problem occurs.
// This must leverage the Ack() interface
func (m *manager) next(receiverKey string, e pkgevent.Event) {

	if e == nil {
		//TODO put metric here
		m.logger.Error().Str("receiverKey", receiverKey).Msg("event is nil")
		return
	}

	m.Lock()
	nextFns := m.receiversFn[receiverKey]

	for wid, n := range nextFns {
		subCtx := logs.SubLoggerCtx(e.Context(), m.logger)
		logs.StrToLogCtx(subCtx, "wid", wid)
		logs.StrToLogCtx(subCtx, "receiverKey", receiverKey)
		childEvt, err := e.Clone(subCtx)
		if err != nil {
			e.Nack(err)
		} else {
			go func(fn pkgreceiver.NextFn, evt event.Event) {
				defer func() {
					p := recover()
					if p != nil {
						panicErr := panics.ToError(p)
						log.Ctx(e.Context()).Error().Str("op", "nextRoute").Str("error", panicErr.Error()).
							Str("stackTrace", panicErr.StackTrace()).Msg("A panic has occurred")
					}
				}()
				fn(evt)
			}(n, childEvt)
		}
	}
	m.Unlock()

	e.Ack()
}

func (m *manager) receive(r *receiver, nextFn pkgreceiver.NextFn) error {
	r.Lock()
	r.done = make(chan struct{})
	r.Unlock()

	m.Lock()
	m.receiversFn[m.mapkey(r.name, r.hash)][r.id] = nextFn
	m.Unlock()

	<-r.done

	return nil
}

func (m *manager) stopReceiving(ctx context.Context, r *receiver) error {
	m.Lock()
	delete(m.receiversFn[m.mapkey(r.name, r.hash)], r.id)
	m.Unlock()

	r.Lock()
	defer r.Unlock()
	//BW
	//r.active = false

	if r.done == nil {
		return &NotRegisteredError{}
	}
	//BW
	//close(r.done)

	return nil
}

func (m *manager) UnregisterReceiver(ctx context.Context, pr pkgreceiver.Receiver) error {
	r, ok := pr.(*receiver)

	// NOTE: No locking on simple reads
	if !ok || !r.active {
		return &RegistrationError{
			Message: fmt.Sprintf("receiver not registered %v", ok),
		}
	}

	r.StopReceiving(ctx) // This in turn calls manager.stopreceiving()

	key := m.mapkey(r.name, r.hash)

	m.Lock()
	defer m.Unlock()

	m.receiversCount[key]--

	if m.receiversCount[key] <= 0 {
		r.receiver.StopReceiving(ctx)
		delete(m.receiversCount, key)
		delete(m.receivers, key)
	}

	delete(m.receiversWrapped, r.id)

	//BW
	r.Lock()
	r.active = false
	r.Unlock()

	return nil

}

// === Filters =======================================================

func (m *manager) Filterers() map[string]pkgfilter.NewFilterer {
	m.Lock()
	defer m.Unlock()

	if m.pm == nil {
		return map[string]pkgfilter.NewFilterer{}
	}
	return m.pm.Filterers()
}

func (m *manager) RegisterFilter(
	ctx context.Context, plugin string,
	name string, config interface{},
	tid tenant.Id,
) (pkgfilter.Filterer, error) {

	factory, err := m.pm.Filterer(plugin)
	if err != nil {
		return nil, &RegistrationError{
			Message: "could not get plugin",
			Plugin:  plugin,
			Name:    name,
			Err:     err,
		}
	}

	hash, err := factory.FiltererHash(config)
	if err != nil {
		return nil, &RegistrationError{
			Message: "could not generate filterer hash",
			Plugin:  plugin,
			Name:    name,
			Err:     err,
		}
	}

	key := m.mapkey(name, hash)

	m.Lock()
	defer m.Unlock()

	f, ok := m.filters[key]
	if !ok {
		f, err = factory.NewFilterer(config)
		if err != nil {
			return nil, &RegistrationError{
				Message: "could not create new filterer",
				Plugin:  plugin,
				Name:    name,
				Err:     err,
			}
		}

		m.filters[key] = f
		m.filtersCount[key] = 0
	}

	u, err := uuid.NewRandom()
	if err != nil {
		return nil, &RegistrationError{
			Message: "could not generate unique id",
			Plugin:  plugin,
			Name:    name,
			Err:     err,
		}
	}

	w := &filter{
		id:   u.String(),
		name: name,
		//plugin:  plugin,
		hash:    hash,
		manager: m,

		filterer: f,
		active:   true,
	}

	m.filtersWrapped[w.id] = w
	m.filtersCount[key]++

	return w, nil

}

func (m *manager) Filters() map[string]pkgfilter.Filterer {
	m.Lock()
	defer m.Unlock()

	filters := map[string]pkgfilter.Filterer{}
	for k, v := range m.filtersWrapped {
		filters[k] = v
	}
	return filters
}

func (m *manager) UnregisterFilter(ctx context.Context, pf pkgfilter.Filterer) error {
	f, ok := pf.(*filter)

	// NOTE:  No locking on simple reads
	if !ok || !f.active {
		return &RegistrationError{
			Message: fmt.Sprintf("filter not registered %v", ok),
		}
	}

	key := m.mapkey(f.name, f.hash)

	{
		m.Lock()
		m.filtersCount[key]--

		if m.filtersCount[key] <= 0 {
			delete(m.filtersCount, key)
			delete(m.filters, key)
		}

		delete(m.filtersWrapped, f.id)
		m.Unlock()
	}

	{
		f.Lock()
		f.active = false
		f.Unlock()
	}

	return nil
}

// === Senders =======================================================

func (m *manager) Senderers() map[string]pkgsender.NewSenderer {
	m.Lock()
	defer m.Unlock()

	if m.pm == nil {
		return map[string]pkgsender.NewSenderer{}
	}

	return m.pm.Senderers()
}

func (m *manager) RegisterSender(
	ctx context.Context, plugin string,
	name string, config interface{},
	tid tenant.Id,
) (pkgsender.Sender, error) {

	ns, err := m.pm.Senderer(plugin)
	if err != nil {
		return nil, &RegistrationError{
			Message: "could not get plugin",
			Plugin:  plugin,
			Name:    name,
			Err:     err,
		}
	}

	hash, err := ns.SenderHash(config)
	if err != nil {
		return nil, &RegistrationError{
			Message: "could not generate hash",
			Plugin:  plugin,
			Name:    name,
			Err:     err,
		}
	}

	key := m.mapkey(name, hash)

	m.Lock()
	defer m.Unlock()

	s, ok := m.senders[key]
	if !ok {
		s, err = ns.NewSender(config)
		if err != nil {
			return nil, &RegistrationError{
				Message: "could not create sender",
				Plugin:  plugin,
				Name:    name,
				Err:     err,
			}
		}

		m.senders[key] = s
		m.sendersCount[key] = 0
	}

	u, err := uuid.NewRandom()
	if err != nil {
		return nil, &RegistrationError{
			Message: "could not generate unique id",
			Plugin:  plugin,
			Name:    name,
			Err:     err,
		}
	}

	w := &sender{
		id:      u.String(),
		name:    name,
		plugin:  plugin,
		hash:    hash,
		manager: m,
		sender:  s,
		active:  true,
	}

	m.sendersWrapped[w.id] = w
	m.sendersCount[key]++

	return w, nil
}

func (m *manager) Senders() map[string]pkgsender.Sender {
	m.Lock()
	defer m.Unlock()
	senders := map[string]pkgsender.Sender{}
	for k, v := range m.sendersWrapped {
		senders[k] = v
	}
	return senders
}

func (m *manager) UniqueSenders() map[string]pkgsender.Sender {
	m.Lock()
	defer m.Unlock()
	senders := map[string]pkgsender.Sender{}
	for _, v := range m.sendersWrapped {
		senders[v.hash] = v
	}
	return senders
}

func (m *manager) UnregisterSender(ctx context.Context, ps pkgsender.Sender) error {

	s, ok := ps.(*sender)
	if !ok || !s.active {
		return &RegistrationError{
			Message: "sender not registered",
		}
	}

	key := m.mapkey(s.name, s.hash)

	m.Lock()

	m.sendersCount[key]--

	if m.sendersCount[key] <= 0 {
		delete(m.sendersCount, key)
		delete(m.senders, key)
	}

	delete(m.sendersWrapped, s.id)
	m.Unlock()

	s.Lock()
	s.active = false
	s.Unlock()

	return nil

}

// === Helper Functions ==============================================

func (m *manager) mapkey(name string, hash string) string {
	return name + "/" + hash
}
