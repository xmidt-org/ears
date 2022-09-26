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

//go:build integration
// +build integration

package gears_test

import (
	"context"
	"github.com/rs/zerolog"
	"github.com/xmidt-org/ears/pkg/tenant"
	"os"
	"testing"
	"time"

	. "github.com/onsi/gomega"
	"github.com/xmidt-org/ears/pkg/event"
	"github.com/xmidt-org/ears/pkg/plugins/gears"
)

func TestGearsSenderReceiver(t *testing.T) {
	logger := zerolog.New(os.Stdout).With().Timestamp().Logger()
	event.SetEventLogger(&logger)
	totalTimeout := 10 * time.Second
	caseTimeout := 5 * time.Second
	testCases := []struct {
		name           string
		timeout        time.Duration
		senderConfig   gears.SenderConfig
		receiverConfig gears.ReceiverConfig
		numMessages    int
	}{
		{
			name:           "one",
			timeout:        caseTimeout,
			senderConfig:   gears.SenderConfig{},
			receiverConfig: gears.ReceiverConfig{},
			numMessages:    5,
		},
	}
	ctx, cancel := context.WithTimeout(context.Background(), totalTimeout)
	defer cancel()
	a := NewWithT(t)
	gearsPlugin, err := gears.NewPlugin()
	a.Expect(err).To(BeNil())
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(ctx, caseTimeout)
			defer cancel()
			a := NewWithT(t)
			// set tc.senderConfig props
			tc.senderConfig = tc.senderConfig.WithDefaults()
			// set tc.receiverConfig props
			tc.receiverConfig.GroupId = "myGroup"
			tc.receiverConfig = tc.receiverConfig.WithDefaults()
			tid := tenant.Id{"myorg", "myapp"}
			gearsReceiver, err := gearsPlugin.NewReceiver(tid, "gears", "gears", tc.receiverConfig, nil)
			a.Expect(err).To(BeNil())
			events := []event.Event{}
			go func() {
				err := gearsReceiver.Receive(func(e event.Event) {
					events = append(events, e)
					e.Ack()
				})
				a.Expect(err).To(BeNil())
				time.Sleep(500 * time.Millisecond)
				a.Expect(events).To(HaveLen(tc.numMessages))
			}()
			sender, err := gearsPlugin.NewSender(tid, "gears", "gears", tc.senderConfig, nil)
			a.Expect(err).To(BeNil())
			for i := 0; i < tc.numMessages; i++ {
				e, err := event.New(ctx, tc.name, event.FailOnNack(t))
				a.Expect(err).To(BeNil())
				sender.Send(e)
			}
			time.Sleep(500 * time.Millisecond)
			gearsSender, ok := sender.(*gears.Sender)
			a.Expect(ok).To(BeTrue())
			a.Expect(gearsSender.Count()).To(BeIdenticalTo(tc.numMessages))
			err = gearsReceiver.StopReceiving(ctx)
			a.Expect(err).To(BeNil())
			gearsSender.StopSending(ctx)
		})
	}
}
