package sqs_test

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

//

import (
	"context"
	"testing"
	"time"

	. "github.com/onsi/gomega"
	"github.com/xmidt-org/ears/pkg/event"
	"github.com/xmidt-org/ears/pkg/plugins/sqs"
)

func TestSender(t *testing.T) {
	totalTimeout := 10 * time.Second
	caseTimeout := 5 * time.Second
	testCases := []struct {
		name           string
		timeout        time.Duration
		senderConfig   sqs.SenderConfig
		receiverConfig sqs.ReceiverConfig
		numMessages    int
	}{
		{
			name:           "one",
			timeout:        caseTimeout,
			senderConfig:   sqs.SenderConfig{},
			receiverConfig: sqs.ReceiverConfig{},
			numMessages:    3,
		},
	}
	ctx, cancel := context.WithTimeout(context.Background(), totalTimeout)
	defer cancel()
	a := NewWithT(t)
	sqsPlugin, err := sqs.NewPlugin()
	a.Expect(err).To(BeNil())
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(ctx, caseTimeout)
			defer cancel()
			a := NewWithT(t)
			tc.senderConfig.QueueUrl = "https://sqs.us-west-2.amazonaws.com/447701116110/ears-test"
			tc.senderConfig = tc.senderConfig.WithDefaults()
			tc.receiverConfig.QueueUrl = "https://sqs.us-west-2.amazonaws.com/447701116110/ears-test"
			one := 1
			tc.receiverConfig.MaxNumberOfMessages = &one
			tc.receiverConfig.WaitTimeSeconds = &one
			tc.receiverConfig.VisibilityTimeout = &one
			tc.receiverConfig.AcknowledgeTimeout = &one
			tc.receiverConfig = tc.receiverConfig.WithDefaults()
			sender, err := sqsPlugin.NewSender(tc.senderConfig)
			a.Expect(err).To(BeNil())
			for i := 0; i < tc.numMessages; i++ {
				e, err := event.New(ctx, tc.name, event.FailOnNack(t))
				a.Expect(err).To(BeNil())
				sender.Send(e)
			}
			sqsSender, ok := sender.(*sqs.Sender)
			a.Expect(ok).To(BeTrue())
			a.Expect(sqsSender.Count()).To(BeIdenticalTo(tc.numMessages))
			sqsReceiver, err := sqsPlugin.NewReceiver(tc.receiverConfig)
			a.Expect(err).To(BeNil())
			events := []event.Event{}
			go func() {
				err = sqsReceiver.Receive(func(e event.Event) {
					events = append(events, e)
					e.Ack()
				})
				a.Expect(err).To(BeNil())
			}()
			time.Sleep(2 * time.Second)
			a.Expect(events).To(HaveLen(tc.numMessages))
			sqsReceiver.StopReceiving(ctx)
		})
	}
}
