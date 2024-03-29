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

// +build integration

package sqs_test

import (
	"context"
	"github.com/rs/zerolog"
	"github.com/xmidt-org/ears/pkg/tenant"
	"os"
	"sync"
	"testing"
	"time"

	. "github.com/onsi/gomega"
	"github.com/xmidt-org/ears/pkg/event"
	"github.com/xmidt-org/ears/pkg/plugins/sqs"
)

func TestSQSSenderReceiver(t *testing.T) {
	logger := zerolog.New(os.Stdout).With().Timestamp().Logger()
	event.SetEventLogger(&logger)
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
			numMessages:    5,
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
			zero := 0
			tc.receiverConfig.MaxNumberOfMessages = &one
			tc.receiverConfig.WaitTimeSeconds = &one
			tc.receiverConfig.VisibilityTimeout = &one
			tc.receiverConfig.AcknowledgeTimeout = &one
			tc.receiverConfig.NumRetries = &zero
			tc.receiverConfig = tc.receiverConfig.WithDefaults()
			tid := tenant.Id{"myorg", "myapp"}
			sender, err := sqsPlugin.NewSender(tid, "sqs", "sqs", tc.senderConfig, nil)
			a.Expect(err).To(BeNil())
			for i := 0; i < tc.numMessages; i++ {
				e, err := event.New(ctx, tc.name, event.FailOnNack(t))
				a.Expect(err).To(BeNil())
				sender.Send(e)
			}
			time.Sleep(2 * time.Second)
			sqsSender, ok := sender.(*sqs.Sender)
			a.Expect(ok).To(BeTrue())
			a.Expect(sqsSender.Count()).To(BeIdenticalTo(tc.numMessages))
			sqsReceiver, err := sqsPlugin.NewReceiver(tid, "sqs", "sqs", tc.receiverConfig, nil)
			a.Expect(err).To(BeNil())
			events := []event.Event{}
			var mutex = &sync.Mutex{}
			go func() {
				err = sqsReceiver.Receive(func(e event.Event) {
					mutex.Lock()
					events = append(events, e)
					mutex.Unlock()
					e.Ack()
				})
				a.Expect(err).To(BeNil())
			}()
			time.Sleep(2 * time.Second)
			mutex.Lock()
			a.Expect(events).To(HaveLen(tc.numMessages))
			mutex.Unlock()
			err = sqsReceiver.StopReceiving(ctx)
			a.Expect(err).To(BeNil())
			sqsSender.StopSending(ctx)
		})
	}
}
