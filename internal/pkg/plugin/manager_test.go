package plugin_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/xmidt-org/ears/internal/pkg/plugin"
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

	t.Fail()
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
	t.Fail()
}

func TestFilterLifecycle(t *testing.T) {
	t.Fail()
}

// === Sender =========================================

func TestSenderRegisterErrors(t *testing.T) {
	t.Fail()
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
	t.Fail()
}

func TestSenderLifecycle(t *testing.T) {
	t.Fail()
}

// === Receiver =========================================

func TestReceiverRegisterErrors(t *testing.T) {
	t.Fail()
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
	t.Fail()
}

func TestReceiverLifecycle(t *testing.T) {
	t.Fail()
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
	pkgfilter.NewFiltererMock
	filterFn filterFn
	events   []pkgevent.Event
}

func (m *newFiltererPluginMock) Name() string    { return "newFiltererPluginMock" }
func (m *newFiltererPluginMock) Version() string { return "version" }
func (m *newFiltererPluginMock) Config() string  { return "config" }

func (m *newFiltererPluginMock) SetFilter(fn filterFn) {
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

				mock.events = append(mock.events, e)
				return mock.filterFn(ctx, e)
			},
		}, nil
	}

	return &mock

}

// === SENDERER ==========================================================

type newSendererPluginMock struct {
	pkgsender.NewSendererMock
	events []pkgevent.Event
}

func (m *newSendererPluginMock) Name() string    { return "newSendererPluginMock" }
func (m *newSendererPluginMock) Version() string { return "version" }
func (m *newSendererPluginMock) Config() string  { return "config" }

func newSenderPlugin(t *testing.T) pkgplugin.Pluginer {
	mock := newSendererPluginMock{}

	mock.SenderHashFunc = func(config interface{}) (string, error) {
		return "sender_" + hasher.Hash(config), nil
	}

	mock.NewSenderFunc = func(config interface{}) (pkgsender.Sender, error) {
		return &pkgsender.SenderMock{
			SendFunc: func(ctx context.Context, e pkgevent.Event) error {
				fmt.Printf("EVENT SENT: %+v\n", e)
				mock.events = append(mock.events, e)
				return nil
			},
		}, nil
	}

	return &mock

}

// === RECEIVER PLUGIN ==========================

type newReceivererPluginMock struct {
	pkgreceiver.NewReceivererMock
	nextFn pkgreceiver.NextFn
	done   chan struct{}
}

func (m *newReceivererPluginMock) Name() string    { return "newReceivererPluginMock" }
func (m *newReceivererPluginMock) Version() string { return "version" }
func (m *newReceivererPluginMock) Config() string  { return "config" }

func (m *newReceivererPluginMock) ReceiveEvent(ctx context.Context, e pkgevent.Event) error {
	return m.nextFn(ctx, e)
}

func newReceiverPlugin(t *testing.T) pkgplugin.Pluginer {
	plugMock := newReceivererPluginMock{
		done: make(chan struct{}),
	}

	receiverMock := pkgreceiver.ReceiverMock{}
	receiverMock.ReceiveFunc = func(ctx context.Context, next receiver.NextFn) error {
		plugMock.nextFn = next
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-plugMock.done:
		}

		return nil
	}

	receiverMock.StopReceivingFunc = func(ctx context.Context) error {
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
