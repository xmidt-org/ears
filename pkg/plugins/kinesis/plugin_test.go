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

package kinesis_test

import (
	"context"
	"github.com/rs/zerolog"
	"github.com/xmidt-org/ears/pkg/plugins/kinesis"
	"github.com/xmidt-org/ears/pkg/tenant"
	"os"
	"sync"
	"testing"
	"time"

	. "github.com/onsi/gomega"
	"github.com/xmidt-org/ears/pkg/event"
)

func TestKinesisSenderReceiver(t *testing.T) {
	logger := zerolog.New(os.Stdout).With().Timestamp().Logger()
	event.SetEventLogger(&logger)
	totalTimeout := 10 * time.Second
	caseTimeout := 5 * time.Second
	testCases := []struct {
		name           string
		timeout        time.Duration
		senderConfig   kinesis.SenderConfig
		receiverConfig kinesis.ReceiverConfig
		numMessages    int
	}{
		{
			name:           "one",
			timeout:        caseTimeout,
			senderConfig:   kinesis.SenderConfig{},
			receiverConfig: kinesis.ReceiverConfig{},
			numMessages:    5,
		},
	}
	ctx, cancel := context.WithTimeout(context.Background(), totalTimeout)
	defer cancel()
	a := NewWithT(t)
	kinesisPlugin, err := kinesis.NewPlugin()
	a.Expect(err).To(BeNil())
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(ctx, caseTimeout)
			defer cancel()
			a := NewWithT(t)
			tc.senderConfig.StreamName = "bw-ears-demo"
			tc.senderConfig = tc.senderConfig.WithDefaults()
			tc.receiverConfig.StreamName = "bw-ears-demo"
			tc.receiverConfig = tc.receiverConfig.WithDefaults()
			tid := tenant.Id{}
			kinesisReceiver, err := kinesisPlugin.NewReceiver(tid, "kinesis", "kinesis", tc.receiverConfig, nil)
			a.Expect(err).To(BeNil())
			events := []event.Event{}
			var mutex = &sync.Mutex{}
			go func() {
				err = kinesisReceiver.Receive(func(e event.Event) {
					mutex.Lock()
					events = append(events, e)
					mutex.Unlock()
					e.Ack()
				})
				a.Expect(err).To(BeNil())
			}()
			time.Sleep(2 * time.Second)
			sender, err := kinesisPlugin.NewSender(tid, "kinesis", "kinesis", tc.senderConfig, nil)
			a.Expect(err).To(BeNil())
			for i := 0; i < tc.numMessages; i++ {
				e, err := event.New(ctx, tc.name, event.FailOnNack(t))
				a.Expect(err).To(BeNil())
				sender.Send(e)
			}
			time.Sleep(2 * time.Second)
			kinesisSender, ok := sender.(*kinesis.Sender)
			a.Expect(ok).To(BeTrue())
			a.Expect(kinesisSender.Count()).To(BeIdenticalTo(tc.numMessages))
			time.Sleep(2 * time.Second)
			mutex.Lock()
			a.Expect(events).To(HaveLen(tc.numMessages))
			mutex.Unlock()
			err = kinesisReceiver.StopReceiving(ctx)
			a.Expect(err).To(BeNil())
			kinesisSender.StopSending(ctx)
		})
	}
}
