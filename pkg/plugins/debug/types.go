package debug

import "sync"

type Receiver struct {
	IntervalMs int
	Rounds     int
	Payload    interface{}

	sync.Mutex
	done chan struct{}
}
