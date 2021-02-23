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
		ctx, _ := context.WithTimeout(e.ctx, ackTO)
		e.ack = ack.NewAckTree(ctx, func() {
			//ok
		}, func(err error) {
			t.Error(err)
		})
		return nil
	}
}

func FailOnAck(t *testing.T) EventOption {
	return func(e *event) error {
		ctx, _ := context.WithTimeout(e.ctx, ackTO)
		e.ack = ack.NewAckTree(ctx, func() {
			t.Error("expecting error")
		}, func(err error) {
			//ok
		})
		return nil
	}
}
