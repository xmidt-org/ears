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

package plugin_test

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"

	"github.com/xmidt-org/ears/internal/pkg/plugin"
	"github.com/xmidt-org/ears/pkg/bit"
	pkgevent "github.com/xmidt-org/ears/pkg/event"
	pkgfilter "github.com/xmidt-org/ears/pkg/filter"
	"github.com/xmidt-org/ears/pkg/hasher"
	pkgplugin "github.com/xmidt-org/ears/pkg/plugin"
	pkgmanager "github.com/xmidt-org/ears/pkg/plugin/manager"
	"github.com/xmidt-org/ears/pkg/receiver"
	pkgreceiver "github.com/xmidt-org/ears/pkg/receiver"
	pkgsender "github.com/xmidt-org/ears/pkg/sender"

	. "github.com/onsi/gomega"
)

// === Filter =========================================
func TestFilterRegisterErrors(t *testing.T) {

	testCases := []struct {
		id     string
		name   string
		plugin string
		config interface{}
		err    *plugin.RegistrationError
	}{
		{
			id:     "no-plugin",
			plugin: "no-plugin",
			name:   "filter-1",
			config: nil,
			err: &plugin.RegistrationError{
				Message: "could not get plugin",
			},
		},
	}

	ctx := context.Background()
	m := newManager(t)
	for _, tc := range testCases {
		t.Run(tc.id, func(t *testing.T) {
			a := NewWithT(t)
			_, err := m.RegisterFilter(ctx, tc.plugin, tc.name, tc.config)
			var re *plugin.RegistrationError
			if errors.As(err, &re) {
				// TODO: Once we start cueing off of messages, it's time to
				// make a new error
				a.Expect(re.Message).To(Equal(tc.err.Message))

			} else {
				t.Error(fmt.Errorf("expected error to be a RegistrationError"))
			}

		})
	}

}

func TestFilterRegister(t *testing.T) {
	ctx := context.Background()
	a := NewWithT(t)

	m := newManager(t)

	fMap := m.Filters()
	a.Expect(len(fMap)).To(Equal(0))

	_, err := m.RegisterFilter(ctx, "filter", "testfilter-1", "noconfig")
	a.Expect(err).To(BeNil())

	fMap = m.Filters()
	a.Expect(len(fMap)).To(Equal(1))

	_, err = m.RegisterFilter(ctx, "filter", "testfilter-2", "noconfig")
	a.Expect(err).To(BeNil())

	fMap = m.Filters()
	a.Expect(len(fMap)).To(Equal(2))
}

func TestFilterUnregister(t *testing.T) {
	ctx := context.Background()
	a := NewWithT(t)

	m := newManager(t)

	fMap := m.Filters()
	a.Expect(len(fMap)).To(Equal(0))

	f, err := m.RegisterFilter(ctx, "filter", "testfilter-1", "noconfig")
	a.Expect(err).To(BeNil())

	fMap = m.Filters()
	a.Expect(len(fMap)).To(Equal(1))

	err = m.UnregisterFilter(ctx, f)
	a.Expect(err).To(BeNil())

	fMap = m.Filters()
	a.Expect(len(fMap)).To(Equal(0))

}

func TestFilterLifecycle(t *testing.T) {
	ctx := context.Background()
	a := NewWithT(t)

	m := newManager(t)

	fMap := m.Filters()
	a.Expect(len(fMap)).To(Equal(0))

	// Register the same filter twice, will only have one unique
	f1, err := m.RegisterFilter(ctx, "filter", "testfilter-1", "noconfig")
	a.Expect(err).To(BeNil())

	fMap = m.Filters()
	a.Expect(len(fMap)).To(Equal(1))

	f2, err := m.RegisterFilter(ctx, "filter", "testfilter-1", "noconfig")
	a.Expect(err).To(BeNil())

	fMap = m.Filters()
	a.Expect(len(fMap)).To(Equal(2))

	err = m.UnregisterFilter(ctx, f1)
	a.Expect(err).To(BeNil())

	fMap = m.Filters()
	a.Expect(len(fMap)).To(Equal(1))

	err = m.UnregisterFilter(ctx, f2)
	a.Expect(err).To(BeNil())

	fMap = m.Filters()
	a.Expect(len(fMap)).To(Equal(0))

}

// === Sender =========================================

func TestSenderRegisterErrors(t *testing.T) {

	testCases := []struct {
		id     string
		name   string
		plugin string
		config interface{}
		err    *plugin.RegistrationError
	}{
		{
			id:     "no-plugin",
			plugin: "no-plugin",
			name:   "sender-1",
			config: nil,
			err: &plugin.RegistrationError{
				Message: "could not get plugin",
			},
		},
	}

	ctx := context.Background()
	m := newManager(t)
	for _, tc := range testCases {
		t.Run(tc.id, func(t *testing.T) {
			a := NewWithT(t)
			_, err := m.RegisterSender(ctx, tc.plugin, tc.name, tc.config)
			var re *plugin.RegistrationError
			if errors.As(err, &re) {
				// TODO: Once we start cueing off of messages, it's time to
				// make a new error
				a.Expect(re.Message).To(Equal(tc.err.Message))

			} else {
				t.Error(fmt.Errorf("expected error to be a RegistrationError"))
			}

		})
	}

}

func TestSenderRegister(t *testing.T) {
	ctx := context.Background()
	a := NewWithT(t)

	m := newManager(t)

	sMap := m.Senders()
	a.Expect(len(sMap)).To(Equal(0))

	_, err := m.RegisterSender(ctx, "sender", "testsender-1", "noconfig")
	a.Expect(err).To(BeNil())

	sMap = m.Senders()
	a.Expect(len(sMap)).To(Equal(1))

	_, err = m.RegisterSender(ctx, "sender", "testsender-2", "noconfig")
	a.Expect(err).To(BeNil())

	sMap = m.Senders()
	a.Expect(len(sMap)).To(Equal(2))
}

func TestSenderUnregister(t *testing.T) {

	ctx := context.Background()
	a := NewWithT(t)

	m := newManager(t)

	sMap := m.Senders()
	a.Expect(len(sMap)).To(Equal(0))

	s, err := m.RegisterSender(ctx, "sender", "testsender-1", "noconfig")
	a.Expect(err).To(BeNil())

	sMap = m.Senders()
	a.Expect(len(sMap)).To(Equal(1))

	err = m.UnregisterSender(ctx, s)
	a.Expect(err).To(BeNil())

	sMap = m.Senders()
	a.Expect(len(sMap)).To(Equal(0))

}

func TestSenderLifecycle(t *testing.T) {

	ctx := context.Background()
	a := NewWithT(t)

	m := newManager(t)

	sMap := m.Senders()
	a.Expect(len(sMap)).To(Equal(0))

	// Register the same sender twice
	s1, err := m.RegisterSender(ctx, "sender", "testsender-1", "noconfig")
	a.Expect(err).To(BeNil())

	sMap = m.Senders()
	a.Expect(len(sMap)).To(Equal(1))

	s2, err := m.RegisterSender(ctx, "sender", "testsender-1", "noconfig")
	a.Expect(err).To(BeNil())

	sMap = m.Senders()
	a.Expect(len(sMap)).To(Equal(2))

	err = m.UnregisterSender(ctx, s1)
	a.Expect(err).To(BeNil())

	sMap = m.Senders()
	a.Expect(len(sMap)).To(Equal(1))

	err = m.UnregisterSender(ctx, s2)
	a.Expect(err).To(BeNil())

	sMap = m.Senders()
	a.Expect(len(sMap)).To(Equal(0))

}

// === Receiver =========================================

func TestReceiverRegisterErrors(t *testing.T) {

	testCases := []struct {
		id     string
		name   string
		plugin string
		config interface{}
		err    *plugin.RegistrationError
	}{
		{
			id:     "no-plugin",
			plugin: "no-plugin",
			name:   "receiver-1",
			config: nil,
			err: &plugin.RegistrationError{
				Message: "could not get plugin",
			},
		},
	}

	ctx := context.Background()
	m := newManager(t)
	for _, tc := range testCases {
		t.Run(tc.id, func(t *testing.T) {
			a := NewWithT(t)
			_, err := m.RegisterReceiver(ctx, tc.plugin, tc.name, tc.config)
			var re *plugin.RegistrationError
			if errors.As(err, &re) {
				// TODO: Once we start cueing off of messages, it's time to
				// make a new error
				a.Expect(re.Message).To(Equal(tc.err.Message))

			} else {
				t.Error(fmt.Errorf("expected error to be a RegistrationError"))
			}

		})
	}

}

func TestReceiverRegister(t *testing.T) {
	ctx := context.Background()
	a := NewWithT(t)

	m := newManager(t)

	rMap := m.Receivers()
	a.Expect(len(rMap)).To(Equal(0))

	_, err := m.RegisterReceiver(ctx, "receiver", "testreceiver-1", "noconfig")
	a.Expect(err).To(BeNil())

	rMap = m.Receivers()
	a.Expect(len(rMap)).To(Equal(1))

	_, err = m.RegisterReceiver(ctx, "receiver", "testreceiver-2", "noconfig")
	a.Expect(err).To(BeNil())

	rMap = m.Receivers()
	a.Expect(len(rMap)).To(Equal(2))

}

func TestReceiverUnregister(t *testing.T) {

	ctx := context.Background()
	a := NewWithT(t)

	m := newManager(t)

	rMap := m.Receivers()
	a.Expect(len(rMap)).To(Equal(0))

	r, err := m.RegisterReceiver(ctx, "receiver", "testreceiver-1", "noconfig")
	a.Expect(err).To(BeNil())

	rMap = m.Receivers()
	a.Expect(len(rMap)).To(Equal(1))

	err = m.UnregisterReceiver(ctx, r)
	a.Expect(err).To(BeNil())

	rMap = m.Receivers()
	a.Expect(len(rMap)).To(Equal(0))

}

func TestReceiverLifecycle(t *testing.T) {

	ctx := context.Background()
	a := NewWithT(t)

	m := newManager(t)

	rMap := m.Receivers()
	a.Expect(len(rMap)).To(Equal(0))

	// Register the same receiver twice
	r1, err := m.RegisterReceiver(ctx, "receiver", "testreceiver-1", "noconfig")
	a.Expect(err).To(BeNil())

	rMap = m.Receivers()
	a.Expect(len(rMap)).To(Equal(1))

	r2, err := m.RegisterReceiver(ctx, "receiver", "testreceiver-1", "noconfig")
	a.Expect(err).To(BeNil())

	rMap = m.Receivers()
	a.Expect(len(rMap)).To(Equal(2))

	next := func(ctx context.Context, e pkgevent.Event) error {
		return nil
	}

	done := make(chan struct{})

	go func() {
		err = m.UnregisterReceiver(ctx, r1)
		a.Expect(err).To(BeNil())

		rMap = m.Receivers()
		a.Expect(len(rMap)).To(Equal(1))

		err = m.UnregisterReceiver(ctx, r2)
		a.Expect(err).To(BeNil())

		rMap = m.Receivers()
		a.Expect(len(rMap)).To(Equal(0))

		close(done)
	}()

	var wg sync.WaitGroup

	for i, r := range []pkgreceiver.Receiver{r1, r2} {
		wg.Add(1)
		go func(r pkgreceiver.Receiver, i int) {
			r.Receive(ctx, next)
			wg.Done()
		}(r, i)
	}

	<-done
	wg.Wait()

}

// === Helper Methods =========================================

func newManager(t *testing.T) plugin.Manager {

	m, err := plugin.NewManager(
		plugin.WithPluginManager(newPluginManager(t)),
	)
	if err != nil {
		t.Error(fmt.Errorf("could not create new manager: %w", err))
	}

	return m

}

func newPluginManager(t *testing.T) pkgmanager.Manager {

	m, err := pkgmanager.New()
	if err != nil {
		t.Error(err)
	}

	m.RegisterPlugin("sender", newSenderPlugin(t))
	m.RegisterPlugin("receiver", newReceiverPlugin(t))
	m.RegisterPlugin("filter", newFiltererPlugin(t))

	return m
}

// === FILTERER ==========================================================

type filterFn func(ctx context.Context, e pkgevent.Event) ([]pkgevent.Event, error)
type newFiltererPluginMock struct {
	sync.Mutex

	pkgfilter.NewFiltererMock
	filterFn filterFn
	events   []pkgevent.Event
}

func (m *newFiltererPluginMock) Name() string     { return "newFiltererPluginMock" }
func (m *newFiltererPluginMock) Version() string  { return "filterVersion" }
func (m *newFiltererPluginMock) Config() string   { return "filterConfig" }
func (m *newFiltererPluginMock) CommitID() string { return "filterCommitID" }
func (m *newFiltererPluginMock) SupportedTypes() bit.Mask {
	return pkgplugin.TypeFilter | pkgplugin.TypePluginer
}

func (m *newFiltererPluginMock) SetFilter(fn filterFn) {
	m.Lock()
	defer m.Unlock()
	m.filterFn = fn
}

func newFiltererPlugin(t *testing.T) pkgplugin.Pluginer {
	mock := newFiltererPluginMock{
		events: []pkgevent.Event{},
		filterFn: func(ctx context.Context, e pkgevent.Event) ([]pkgevent.Event, error) {
			return []pkgevent.Event{e}, nil
		},
	}

	mock.FiltererHashFunc = func(config interface{}) (string, error) {
		return "filter_" + hasher.Hash(config), nil
	}

	mock.NewFiltererFunc = func(config interface{}) (pkgfilter.Filterer, error) {
		return &pkgfilter.FiltererMock{
			FilterFunc: func(ctx context.Context, e pkgevent.Event) ([]pkgevent.Event, error) {
				fmt.Printf("FILTER EVENT: %+v\n", e)

				mock.Lock()
				defer mock.Unlock()
				mock.events = append(mock.events, e)
				return mock.filterFn(ctx, e)
			},
		}, nil
	}

	return &mock

}

// === SENDERER ==========================================================

type newSendererPluginMock struct {
	sync.Mutex
	pkgsender.NewSendererMock
	events []pkgevent.Event
}

func (m *newSendererPluginMock) Name() string     { return "newSendererPluginMock" }
func (m *newSendererPluginMock) Version() string  { return "senderVersion" }
func (m *newSendererPluginMock) Config() string   { return "senderConfig" }
func (m *newSendererPluginMock) CommitID() string { return "senderCommitID" }
func (m *newSendererPluginMock) SupportedTypes() bit.Mask {
	return pkgplugin.TypeSender | pkgplugin.TypePluginer
}

func newSenderPlugin(t *testing.T) pkgplugin.Pluginer {
	mock := newSendererPluginMock{}

	mock.SenderHashFunc = func(config interface{}) (string, error) {
		return "sender_" + hasher.Hash(config), nil
	}

	mock.NewSenderFunc = func(config interface{}) (pkgsender.Sender, error) {
		return &pkgsender.SenderMock{
			SendFunc: func(ctx context.Context, e pkgevent.Event) error {
				fmt.Printf("EVENT SENT: %+v\n", e)
				mock.Lock()
				defer mock.Unlock()
				mock.events = append(mock.events, e)
				return nil
			},
		}, nil
	}

	return &mock

}

// === RECEIVER PLUGIN ==========================

type newReceivererPluginMock struct {
	sync.Mutex
	pkgreceiver.NewReceivererMock
	nextFn pkgreceiver.NextFn
	done   chan struct{}
}

func (m *newReceivererPluginMock) Name() string     { return "newReceivererPluginMock" }
func (m *newReceivererPluginMock) Version() string  { return "senderVersion" }
func (m *newReceivererPluginMock) Config() string   { return "senderConfig" }
func (m *newReceivererPluginMock) CommitID() string { return "senderCommitID" }
func (m *newReceivererPluginMock) SupportedTypes() bit.Mask {
	return pkgplugin.TypeReceiver | pkgplugin.TypePluginer
}

func (m *newReceivererPluginMock) ReceiveEvent(ctx context.Context, e pkgevent.Event) error {
	return m.nextFn(ctx, e)
}

func newReceiverPlugin(t *testing.T) pkgplugin.Pluginer {
	plugMock := newReceivererPluginMock{
		done: make(chan struct{}),
	}

	receiverMock := pkgreceiver.ReceiverMock{}
	receiverMock.ReceiveFunc = func(ctx context.Context, next receiver.NextFn) error {
		plugMock.Lock()

		plugMock.nextFn = next
		plugMock.Unlock()

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-plugMock.done:
		}

		return nil
	}

	receiverMock.StopReceivingFunc = func(ctx context.Context) error {
		plugMock.Lock()
		defer plugMock.Unlock()

		close(plugMock.done)
		return nil
	}

	plugMock.ReceiverHashFunc = func(config interface{}) (string, error) {
		return "receiver_" + hasher.Hash(config), nil
	}

	plugMock.NewReceiverFunc = func(config interface{}) (pkgreceiver.Receiver, error) {
		return &receiverMock, nil
	}

	return &plugMock
}
