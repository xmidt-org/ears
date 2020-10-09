/**
 *  Copyright (c) 2020  Comcast Cable Communications Management, LLC
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package internal

/*import "context"

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
}*/
