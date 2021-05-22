// Copyright 2021 Comcast Cable Communications Management, LLC
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

package event

import (
	"context"
	"github.com/xmidt-org/ears/internal/pkg/ack"
	"testing"
	"time"
)

var ackTO = time.Second * 10

func FailOnNack(t *testing.T) EventOption {
	return func(e *event) error {
		ctx, cancel := context.WithTimeout(e.ctx, ackTO)
		e.ack = ack.NewAckTree(ctx, func() {
			cancel()
		}, func(err error) {
			cancel()
			t.Error(err)
		})
		return nil
	}
}

func FailOnAck(t *testing.T) EventOption {
	return func(e *event) error {
		ctx, cancel := context.WithTimeout(e.ctx, ackTO)
		e.ack = ack.NewAckTree(ctx, func() {
			cancel()
			t.Error("expecting error")
		}, func(err error) {
			cancel()
		})
		return nil
	}
}
