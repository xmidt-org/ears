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
	pm pkgmanager.Manager

	receivers        map[string]pkgreceiver.Receiver
	receiversCount   map[string]int
	receiversWrapped map[string]*receiver
	receiversFn      map[string]map[string]pkgreceiver.NextFn // map[receiverKey]map[wrapperID] -> nextFN

	filters      map[string]*filter
	filtersCount map[string]int
	senders      map[string]*sender
	sendersCount map[string]int

	nextFnDeadline time.Duration
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

		filters:      map[string]*filter{},
		filtersCount: map[string]int{},

		senders:      map[string]*sender{},
		sendersCount: map[string]int{},
	}

	var err error
	for _, option := range options {
		err = option(&m)
		if err != nil {
			// TODO: OptionError
			return nil, &OptionError{Err: err}
		}
	}

	return &m, nil
}

func (m *manager) WithPluginManager(p pkgmanager.Manager) error {
	if p == nil {
		return fmt.Errorf("plugin manager cannot be nil")
	}

	m.pm = p
	return nil
}

func (m *manager) WithNextFnDeadline(d time.Duration) error {
	m.nextFnDeadline = d
	return nil
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
	name string, config string,
) (pkgreceiver.Receiver, error) {

	ns, err := m.pm.Receiverer(plugin)
	if err != nil {
		return nil, err
	}

	hash, err := ns.ReceiverHash(config)
	if err != nil {
		return nil, fmt.Errorf("could not generate hash: %w", err)
	}

	receiverKey := m.mapkey(name, hash)

	r, ok := m.receivers[receiverKey]
	if !ok {
		r, err = ns.NewReceiver(config)
		if err != nil {
			return nil, &RegistrationError{Err: fmt.Errorf("could not create new receiver: %w", err)}
		}
		m.receiversCount[receiverKey]++
		m.receiversFn[receiverKey] = map[string]pkgreceiver.NextFn{}

		// TODO: Determine lifecycle
		go func() {
			r.Receive(ctx, func(ctx context.Context, e event.Event) error {
				return m.next(ctx, receiverKey, e)
			})
		}()
	}

	u, err := uuid.NewUUID()
	if err != nil {
		return nil, &RegistrationError{Err: fmt.Errorf("could not generate unique id: %w", err)}
	}

	w := &receiver{
		key:      u.String(),
		name:     name,
		hash:     hash,
		manager:  m,
		receiver: r,
		active:   true,
	}

	m.receiversWrapped[w.key] = w
	m.receiversCount[receiverKey]++

	return r, nil

}

func (m *manager) Receivers() map[string]pkgreceiver.Receiver {
	receivers := map[string]pkgreceiver.Receiver{}
	for k, v := range m.receiversWrapped {
		receivers[k] = v
	}
	return receivers
}

// next iterates through all receiver functions that have registered for
// a receiver (unique by name + config hash).  These must be independent,
// so no error can actually be returned to the receiver if a problem occurs.
// This must leverage the Ack() interface
func (m *manager) next(ctx context.Context, receiverKey string, e pkgevent.Event) error {
	//
	ctx, cancel := context.WithTimeout(ctx, m.nextFnDeadline)
	defer cancel()

	nextFns := m.receiversFn[receiverKey]
	errCh := make(chan error, len(nextFns))

	fmt.Println("MANAGER NEXT CALLED -- SIZE OF NEXTFNS IS ", len(nextFns))

	var wg sync.WaitGroup
	for _, n := range nextFns {
		wg.Add(1)
		go func(fn pkgreceiver.NextFn) {
			err := fn(ctx, e)
			if err != nil {
				errCh <- err
			}
		}(n)
	}

	wg.Wait()
	close(errCh)

	// Does it make sense to return an error if any of the filters fail?
	if len(errCh) > 0 {
		return <-errCh
	}

	return nil

}

func (m *manager) receive(ctx context.Context, r *receiver, nextFn pkgreceiver.NextFn) error {
	r.done = make(chan struct{})

	m.receiversFn[m.mapkey(r.name, r.hash)][r.key] = nextFn

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-r.done:
	}

	return nil
}

func (m *manager) stopreceiving(ctx context.Context, r *receiver) error {
	delete(m.receiversFn[m.mapkey(r.name, r.hash)], r.key)
	close(r.done)
	return nil
}

func (m *manager) UnregisterReceiver(ctx context.Context, r pkgreceiver.Receiver) error {
	r.StopReceiving(ctx) // This in turn calls manager.stopreceiving()
	// TODO
	return &RegistrationError{
		Err: fmt.Errorf("not yet implemented"),
	}
}

// === Filters =======================================================

func (m *manager) Filterers() map[string]pkgfilter.NewFilterer {
	if m.pm == nil {
		return map[string]pkgfilter.NewFilterer{}
	}
	return m.pm.Filterers()
}

func (m *manager) RegisterFilter(
	ctx context.Context, plugin string,
	name string, config string,
) (pkgfilter.Filterer, error) {

	factory, err := m.pm.Filterer(plugin)
	if err != nil {
		return nil, &RegistrationError{
			Err: fmt.Errorf("could not get filterer (plugin=%s): %w", plugin, err),
		}
	}

	hash, err := factory.FiltererHash(config)
	if err != nil {
		return nil, &RegistrationError{
			Err: fmt.Errorf("could not generate filterer hash (plugin=%s): %w", plugin, err),
		}
	}

	key := m.mapkey(name, hash)
	w, ok := m.filters[key]
	if ok {
		m.filtersCount[key]++
		return w.filterer, nil
	}

	f, err := factory.NewFilterer(config)
	if err != nil {
		return nil, fmt.Errorf("could not create filterer (plugin=%s): %w", plugin, err)
	}

	w = &filter{
		id:      key,
		name:    name,
		hash:    hash,
		manager: m,

		filterer: f,
		active:   true,
	}

	m.filters[key] = w
	m.filtersCount[key]++

	return w, nil

}

func (m *manager) Filters() map[string]pkgfilter.Filterer {
	filters := map[string]pkgfilter.Filterer{}
	for k, v := range m.filters {
		filters[k] = v
	}
	return filters
}

func (m *manager) UnregisterFilter(ctx context.Context, f pkgfilter.Filterer) error {
	// TODO:
	return nil
}

// === Senders =======================================================

func (m *manager) Senderers() map[string]pkgsender.NewSenderer {
	if m.pm == nil {
		return map[string]pkgsender.NewSenderer{}
	}

	return m.pm.Senderers()
}

func (m *manager) RegisterSender(
	ctx context.Context, plugin string,
	name string, config string,
) (pkgsender.Sender, error) {

	ns, err := m.pm.Senderer(plugin)
	if err != nil {
		return nil, &RegistrationError{
			Err: fmt.Errorf("could not get senderer (%s): %w", plugin, err),
		}
	}

	hash, err := ns.SenderHash(config)
	if err != nil {
		return nil, &RegistrationError{
			Err: fmt.Errorf("could not generate hash: %w", err),
		}
	}

	key := m.mapkey(name, hash)

	w, ok := m.senders[key]
	if ok {
		return w.sender, nil
	}

	s, err := ns.NewSender(config)
	if err != nil {
		return nil, &RegistrationError{
			Err: fmt.Errorf("could not create sender: %w", err),
		}
	}

	w = &sender{
		id:      key,
		name:    name,
		hash:    hash,
		manager: m,
		sender:  s,
		active:  true,
	}

	m.senders[key] = w
	m.sendersCount[key]++

	return w, nil
}

func (m *manager) Senders() map[string]pkgsender.Sender {
	senders := map[string]pkgsender.Sender{}
	for k, v := range m.senders {
		senders[k] = v
	}
	return senders
}

func (m *manager) UnregisterSender(ctx context.Context, s pkgsender.Sender) error {
	return &RegistrationError{
		Err: fmt.Errorf("not yet implemented"),
	}
}

// === Helper Functions ==============================================

func fnWithTimeout(ctx context.Context, fn func() error) error {

	done := make(chan struct{}, 1)
	var err error

	go func() {
		err = fn()
		close(done)
	}()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-done:
	}

	return err
}

func (m *manager) mapkey(name string, hash string) string {
	return name + "/" + hash
}
