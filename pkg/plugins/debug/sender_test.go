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

package debug_test

import (
	"context"

	"testing"
	"time"

	"github.com/xmidt-org/ears/pkg/plugins/debug"
	"github.com/xorcare/pointer"

	"github.com/xmidt-org/ears/pkg/event"

	. "github.com/onsi/gomega"
)

func TestSender(t *testing.T) {

	totalTimeout := 5 * time.Second
	caseTimeout := 2 * time.Second

	testCases := []struct {
		name        string
		timeout     time.Duration
		numMessages int
		config      debug.SenderConfig
	}{
		{
			name:    "none",
			timeout: caseTimeout,
			config:  debug.SenderConfig{},
		},

		{
			name:        "underflow",
			timeout:     caseTimeout,
			numMessages: 1,
			config: debug.SenderConfig{
				MaxHistory: pointer.Int(5),
			},
		},

		{
			name:        "overflow",
			timeout:     caseTimeout,
			numMessages: 7,
			config: debug.SenderConfig{
				MaxHistory: pointer.Int(3),
			},
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), totalTimeout)
	defer cancel()

	a := NewWithT(t)
	p, err := debug.NewPlugin()
	a.Expect(err).To(BeNil())

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(ctx, caseTimeout)
			defer cancel()

			a := NewWithT(t)

			w := &debug.SendSlice{}
			tc.config.Destination = debug.DestinationCustom
			tc.config.Writer = w
			tc.config = tc.config.WithDefaults()

			s, err := p.NewSender(tid, "debug", "mydebug", tc.config, nil)
			a.Expect(err).To(BeNil())

			for i := 0; i < tc.numMessages; i++ {
				e, err := event.New(ctx, tc.name, event.FailOnNack(t))
				a.Expect(err).To(BeNil())
				s.Send(e)
			}

			a.Expect(w.Events()).To(HaveLen(tc.numMessages))

			ds, ok := s.(*debug.Sender)
			a.Expect(ok).To(BeTrue())

			a.Expect(ds.History()).To(HaveLen(
				min(tc.numMessages, *tc.config.MaxHistory),
			))

		})

	}

}
