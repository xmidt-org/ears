package route_test

import (
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/xmidt-org/ears/pkg/filter"
	"github.com/xmidt-org/ears/pkg/receiver"
	"github.com/xmidt-org/ears/pkg/route"
	"github.com/xmidt-org/ears/pkg/sender"

	. "github.com/onsi/gomega"
)

var defaultTestTimeout = 3 * time.Second

func TestErrorCases(t *testing.T) {
	testCases := []struct {
		name string
		r    receiver.Receiver
		f    filter.Filterer
		s    sender.Sender
		err  error
	}{
		{
			name: "nil receiver",
			r:    nil,
			f:    &filter.FiltererMock{},
			s:    &sender.SenderMock{},
			err:  &route.InvalidRouteError{},
		},
		{
			name: "nil sender",
			r:    &receiver.ReceiverMock{},
			f:    &filter.FiltererMock{},
			s:    nil,
			err:  &route.InvalidRouteError{},
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), defaultTestTimeout)
	defer cancel()

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			a := NewWithT(t)

			err := (&route.Route{}).Run(ctx, tc.r, tc.f, tc.s)
			a.Expect(err).ToNot(BeNil())
			a.Expect(errTypeToString(err)).To(Equal(errTypeToString(tc.err)))

		})
	}
}

func TestRoutes(t *testing.T) {
	testCases := []struct {
		name string
		r    receiver.Receiver
		f    filter.Filterer
		s    sender.Sender
		err  error
	}{
		{
			name: "nil receiver",
			r:    nil,
			f:    &filter.FiltererMock{},
			s:    &sender.SenderMock{},
			err:  &route.InvalidRouteError{},
		},
		{
			name: "nil sender",
			r:    &receiver.ReceiverMock{},
			f:    &filter.FiltererMock{},
			s:    nil,
			err:  &route.InvalidRouteError{},
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), defaultTestTimeout)
	defer cancel()

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			a := NewWithT(t)

			err := (&route.Route{}).Run(ctx, tc.r, tc.f, tc.s)
			a.Expect(err).ToNot(BeNil())
			a.Expect(errTypeToString(err)).To(Equal(errTypeToString(tc.err)))

		})
	}
}

// =========================================================================

func errTypeToString(err error) string {
	if err == nil {
		return "<nil>"
	}

	return reflect.TypeOf(err).String()
}
