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
			r: &receiver.ReceiverMock{
				StopReceivingFunc: func(ctx context.Context) error {
					return nil
				},
			},
			f:   &filter.FiltererMock{},
			s:   nil,
			err: &route.InvalidRouteError{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			a := NewWithT(t)

			err := (&route.Route{}).Run(tc.r, tc.f, tc.s)
			a.Expect(err).ToNot(BeNil())
			a.Expect(errTypeToString(err)).To(Equal(errTypeToString(tc.err)))
		})
		if tc.r != nil {
			tc.r.StopReceiving(context.Background())
		}
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
			r: &receiver.ReceiverMock{
				StopReceivingFunc: func(ctx context.Context) error {
					return nil
				},
			},
			f:   &filter.FiltererMock{},
			s:   nil,
			err: &route.InvalidRouteError{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			a := NewWithT(t)

			err := (&route.Route{}).Run(tc.r, tc.f, tc.s)
			a.Expect(err).ToNot(BeNil())
			a.Expect(errTypeToString(err)).To(Equal(errTypeToString(tc.err)))

		})
		if tc.r != nil {
			tc.r.StopReceiving(context.Background())
		}
	}
}

// =========================================================================

func errTypeToString(err error) string {
	if err == nil {
		return "<nil>"
	}

	return reflect.TypeOf(err).String()
}
