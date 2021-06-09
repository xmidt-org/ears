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
	. "github.com/onsi/gomega"
	"github.com/xmidt-org/ears/pkg/plugins/debug"
	"testing"
)

func TestSendReceive(t *testing.T) {
	a := NewWithT(t)
	p, err := debug.NewPlugin()
	a.Expect(err).To(BeNil())

	r, err := p.NewReceiver(tid, "debug", "mydebug", "")

	a.Expect(err).To(BeNil())

	s, err := p.NewSender("")
	a.Expect(err).To(BeNil())

	err = r.Receive(s.Send)
	a.Expect(err).To(BeNil())
	r.StopReceiving(context.Background())
}
