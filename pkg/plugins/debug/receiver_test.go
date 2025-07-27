// SPDX-FileCopyrightText: 2025 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package debug_test

import (
	"context"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"github.com/spf13/viper"
	"github.com/xmidt-org/ears/internal/pkg/syncer"
	"github.com/xmidt-org/ears/pkg/tenant"

	"github.com/xmidt-org/ears/pkg/plugins/debug"
	"github.com/xorcare/pointer"

	"github.com/xmidt-org/ears/pkg/event"

	. "github.com/onsi/gomega"
)

var tid = tenant.Id{OrgId: "myorg", AppId: "myapp"}

func TestReceiver(t *testing.T) {
	caseTimeout := 2 * time.Second

	testCases := []struct {
		name    string
		timeout time.Duration
		config  debug.ReceiverConfig
	}{
		{
			name:    "none",
			timeout: caseTimeout,
			config: debug.ReceiverConfig{
				Rounds:  pointer.Int(0),
				Payload: "none",
			},
		},

		{
			name:    "one",
			timeout: caseTimeout,
			config: debug.ReceiverConfig{
				Rounds:     pointer.Int(1),
				IntervalMs: pointer.Int(10),
				Payload:    "one",
				MaxHistory: pointer.Int(5),
			},
		},

		{
			name:    "five",
			timeout: caseTimeout,
			config: debug.ReceiverConfig{
				Rounds:     pointer.Int(5),
				IntervalMs: pointer.Int(10),
				Payload:    "five",
				MaxHistory: pointer.Int(3),
			},
		},
	}

	a := NewWithT(t)
	p, err := debug.NewPlugin()
	a.Expect(err).To(BeNil())

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			a := NewWithT(t)

			// Make sure we fill in all values
			tc.config = tc.config.WithDefaults()
			logger := zerolog.Nop()
			r, err := p.NewReceiver(tid, "debug", "mydebug", tc.config, nil, syncer.NewInMemoryDeltaSyncer(&logger, viper.New()))
			a.Expect(err).To(BeNil())

			events := []event.Event{}

			err = r.Receive(func(e event.Event) {
				events = append(events, e)
				e.Ack()
			})

			a.Expect(err).To(BeNil())
			a.Expect(events).To(HaveLen(*tc.config.Rounds))

			dr, ok := r.(*debug.Receiver)
			a.Expect(ok).To(BeTrue())
			a.Expect(dr.Count()).To(Equal(*tc.config.Rounds))

			history := dr.History()

			a.Expect(history).To(HaveLen(
				min(*tc.config.Rounds, *tc.config.MaxHistory),
			))

			for _, e := range history {
				p, ok := e.Payload().(string)
				a.Expect(ok).To(BeTrue())
				a.Expect(p).To(Equal(tc.config.Payload))
			}
			r.StopReceiving(context.Background())
		})

	}

}

func TestReceiveErrors(t *testing.T) {

}
