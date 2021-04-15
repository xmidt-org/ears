package redis_test

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
	"github.com/xmidt-org/ears/pkg/plugins/redis"
	"testing"
	"time"

	. "github.com/onsi/gomega"
	"github.com/xmidt-org/ears/pkg/event"
)

func TestSQSSenderReceiver(t *testing.T) {
	totalTimeout := 10 * time.Second
	caseTimeout := 5 * time.Second
	testCases := []struct {
		name           string
		timeout        time.Duration
		senderConfig   redis.SenderConfig
		receiverConfig redis.ReceiverConfig
		numMessages    int
	}{
		{
			name:           "one",
			timeout:        caseTimeout,
			senderConfig:   redis.SenderConfig{},
			receiverConfig: redis.ReceiverConfig{},
			numMessages:    5,
		},
	}
	ctx, cancel := context.WithTimeout(context.Background(), totalTimeout)
	defer cancel()
	a := NewWithT(t)
	redisPlugin, err := redis.NewPlugin()
	a.Expect(err).To(BeNil())
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(ctx, caseTimeout)
			defer cancel()
			a := NewWithT(t)
			tc.senderConfig.Endpoint = "localhost:6379"
			tc.senderConfig.Channel = "ears_plugin_test"
			tc.senderConfig = tc.senderConfig.WithDefaults()
			tc.receiverConfig.Endpoint = "localhost:6379"
			tc.receiverConfig.Channel = "ears_plugin_test"
			tc.receiverConfig = tc.receiverConfig.WithDefaults()
			redisReceiver, err := redisPlugin.NewReceiver(tc.receiverConfig)
			a.Expect(err).To(BeNil())
			events := []event.Event{}
			go func() {
				err = redisReceiver.Receive(func(e event.Event) {
					events = append(events, e)
					e.Ack()
				})
				a.Expect(err).To(BeNil())
				time.Sleep(200 * time.Millisecond)
				a.Expect(events).To(HaveLen(tc.numMessages))
			}()
			sender, err := redisPlugin.NewSender(tc.senderConfig)
			a.Expect(err).To(BeNil())
			for i := 0; i < tc.numMessages; i++ {
				e, err := event.New(ctx, tc.name, event.FailOnNack(t))
				a.Expect(err).To(BeNil())
				sender.Send(e)
			}
			time.Sleep(200 * time.Millisecond)
			redisSender, ok := sender.(*redis.Sender)
			a.Expect(ok).To(BeTrue())
			a.Expect(redisSender.Count()).To(BeIdenticalTo(tc.numMessages))

			err = redisReceiver.StopReceiving(ctx)
			a.Expect(err).To(BeNil())
			redisSender.StopSending(ctx)
		})
	}
}
