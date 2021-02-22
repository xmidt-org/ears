package event

import (
	"github.com/xmidt-org/ears/internal/pkg/ack"
	"testing"
)

func FailOnNack(t *testing.T) EventOption {
	return func(e *event) error {
		e.ack = ack.NewAckTree(e.ctx, func() {
			//ok
		}, func(err error) {
			t.Error(err)
		})
		return nil
	}
}

func FailOnAck(t *testing.T) EventOption {
	return func(e *event) error {
		e.ack = ack.NewAckTree(e.ctx, func() {
			t.Error("expecting error")
		}, func(err error) {
			//ok
		})
		return nil
	}
}
