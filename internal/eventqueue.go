package internal

import "context"

const (
	EventQueueDepth = 1000
)

type (
	EventQueue struct {
		eventChannel chan *Event
	}
)

var (
	eventQueue EventQueuer
)

func NewEventQueue(ctx context.Context) EventQueuer {
	eq := new(EventQueue)
	eq.eventChannel = make(chan *Event, EventQueueDepth)
	return eq
}

func GetEventQueue(ctx context.Context) EventQueuer {
	if eventQueue == nil {
		eventQueue = NewEventQueue(ctx)
	}
	return eventQueue
}

func (eq *EventQueue) AddEvent(ctx context.Context, event *Event) {
	eq.eventChannel <- event
}

func (eq *EventQueue) NextEvent(ctx context.Context) *Event {
	evt := <-eq.eventChannel
	return evt
}

func (eq *EventQueue) GetEventCount(ctx context.Context) int {
	return len(eq.eventChannel)
}

func (eq *EventQueue) GetMaxEventCount(ctx context.Context) int {
	return EventQueueDepth
}

func (eq *EventQueue) GetChannel(ctx context.Context) chan *Event {
	return eq.eventChannel
}
