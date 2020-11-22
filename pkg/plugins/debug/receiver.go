package debug

import (
	"context"

	"fmt"
	"time"

	"github.com/xmidt-org/ears/pkg/event"
	"github.com/xmidt-org/ears/pkg/receiver"
)

// var _ receiver.NewReceiverer = (*Receiver)(nil)
var _ receiver.Receiver = (*Receiver)(nil)

func NewReceiver(config string) (receiver.Receiver, error) {
	return &Receiver{
		IntervalMs: 1000,
		Rounds:     10,
		Payload:    "debug message",
	}, nil
}

func (r *Receiver) Receive(ctx context.Context, next receiver.NextFn) error {
	if next == nil {
		return &receiver.InvalidConfigError{
			Err: fmt.Errorf("next cannot be nil"),
		}
	}

	r.done = make(chan struct{})

	go func() {
		defer func() {
			r.Lock()
			if r.done != nil {
				close(r.done)
			}
			r.Unlock()
		}()

		for count := 0; count != r.Rounds; count++ {
			select {
			case <-ctx.Done():
				return
			case <-time.After(time.Duration(r.IntervalMs) * time.Millisecond):
				e, err := event.NewEvent(r.Payload)
				if err != nil {
					return
				}

				err = next(ctx, e)
				if err != nil {
					return
				}

			}
		}
	}()

	<-r.done

	return ctx.Err()
}

func (r *Receiver) StopReceiving(ctx context.Context) error {
	r.Lock()
	defer r.Unlock()

	if r.done != nil {
		close(r.done)
	}

	return nil
}
