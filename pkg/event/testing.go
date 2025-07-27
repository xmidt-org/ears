// SPDX-FileCopyrightText: 2025 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package event

import (
	"context"
	"testing"
	"time"

	"github.com/xmidt-org/ears/internal/pkg/ack"
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
