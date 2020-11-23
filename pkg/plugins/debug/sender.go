package debug

import (
	"context"
	"fmt"
	"os"

	"github.com/xmidt-org/ears/pkg/event"
	"github.com/xmidt-org/ears/pkg/sender"
)

func NewSender(config string) (sender.Sender, error) {
	return &Sender{
		Destination: os.Stdout,
	}, nil
}

func (s *Sender) Send(ctx context.Context, e event.Event) error {
	if s.Destination == nil {
		return &sender.InvalidConfigError{
			Err: fmt.Errorf("destination cannot be null"),
		}
	}

	fmt.Fprintln(s.Destination, e)
	return nil
}
