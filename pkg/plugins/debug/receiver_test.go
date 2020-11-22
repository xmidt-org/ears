package debug_test

import (
	"context"
	"fmt"

	"testing"
	"time"

	"github.com/xmidt-org/ears/pkg/plugins/debug"

	"github.com/xmidt-org/ears/pkg/event"
)

func TestReceiver(t *testing.T) {
	r, err := debug.NewReceiver("")
	if err != nil {
		t.Errorf(err.Error())
	}

	ctx, cancel := context.WithTimeout(context.Background(), 12*time.Second)
	defer cancel()

	err = r.Receive(ctx, func(ctx context.Context, e event.Event) error {
		fmt.Printf("was sent a message: %+v\n", e)
		return nil
	})

	if err != nil {
		fmt.Println("receive returned an error: ", err)
	}
}
