package route

import (
	"context"
	"sync"

	"github.com/xmidt-org/ears/pkg/filter"
	"github.com/xmidt-org/ears/pkg/receiver"
	"github.com/xmidt-org/ears/pkg/sender"
)

type Router interface {
	Run(ctx context.Context, r receiver.Receiver, f filter.Filterer, s sender.Sender) error
	Stop(ctx context.Context) error
}

type Route struct {
	sync.Mutex

	r receiver.Receiver
	f filter.Filterer
	s sender.Sender
}

type InvalidRouteError struct {
	Err error
}
