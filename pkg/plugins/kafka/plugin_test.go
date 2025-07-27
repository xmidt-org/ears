// SPDX-FileCopyrightText: 2025 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

//go:build integration
// +build integration

package kafka_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"github.com/xmidt-org/ears/pkg/tenant"

	. "github.com/onsi/gomega"
	"github.com/xmidt-org/ears/pkg/event"
	"github.com/xmidt-org/ears/pkg/plugins/kafka"
)

func TestKafkaSenderReceiver(t *testing.T) {
	logger := zerolog.New(os.Stdout).With().Timestamp().Logger()
	event.SetEventLogger(&logger)
	totalTimeout := 10 * time.Second
	caseTimeout := 5 * time.Second
	testCases := []struct {
		name           string
		timeout        time.Duration
		senderConfig   kafka.SenderConfig
		receiverConfig kafka.ReceiverConfig
		numMessages    int
	}{
		{
			name:           "one",
			timeout:        caseTimeout,
			senderConfig:   kafka.SenderConfig{},
			receiverConfig: kafka.ReceiverConfig{},
			numMessages:    5,
		},
	}
	ctx, cancel := context.WithTimeout(context.Background(), totalTimeout)
	defer cancel()
	a := NewWithT(t)
	kafkaPlugin, err := kafka.NewPlugin()
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
			kafkaReceiver, err := kafkaPlugin.NewReceiver(tid, "kafka", "kafka", tc.receiverConfig, nil)
			a.Expect(err).To(BeNil())
			events := []event.Event{}
			go func() {
				err := kafkaReceiver.Receive(func(e event.Event) {
					events = append(events, e)
					e.Ack()
				})
				a.Expect(err).To(BeNil())
				time.Sleep(500 * time.Millisecond)
				a.Expect(events).To(HaveLen(tc.numMessages))
			}()
			sender, err := kafkaPlugin.NewSender(tid, "kafka", "kafka", tc.senderConfig, nil)
			a.Expect(err).To(BeNil())
			for i := 0; i < tc.numMessages; i++ {
				e, err := event.New(ctx, tc.name, event.FailOnNack(t))
				a.Expect(err).To(BeNil())
				sender.Send(e)
			}
			time.Sleep(500 * time.Millisecond)
			kafkaSender, ok := sender.(*kafka.Sender)
			a.Expect(ok).To(BeTrue())
			a.Expect(kafkaSender.Count()).To(BeIdenticalTo(tc.numMessages))
			err = kafkaReceiver.StopReceiving(ctx)
			a.Expect(err).To(BeNil())
			kafkaSender.StopSending(ctx)
		})
	}
}
