// Licensed to Comcast Cable Communications Management, LLC under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. Comcast Cable Communications Management, LLC licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package event

import "context"

//go:generate rm -f testing_mock.go
//go:generate moq -out testing_mock.go . Event NewEventerer

type Event interface {
	//Get the event payload
	Payload() interface{}

	//Get the event context
	Context() context.Context

	//Set the event payload
	//Will return an error if the event is done
	SetPayload(payload interface{}) error

	//Replace the current event context
	//Will return an error if the event is done
	SetContext(ctx context.Context) error

	//Acknowledge that the event is handled successfully
	//Further actions on the event are no longer possible
	Ack()

	//Acknowledge that there is an error handling the event
	//Further actions on the event are no longer possible
	Nack(err error)

	//Create a child the event with payload deep-copied
	//Clone fails and return an error if the event is already acknowledged
	Clone(ctx context.Context) (Event, error)
}

type NewEventerer interface {

	//Create a new event given a context and a payload
	NewEvent(ctx context.Context, payload interface{}) (Event, error)

	//Create a new event given a context and a payload, and two completion functions,
	//handledFn and errFn. An event constructed this way will be notified through
	//the handledFn when an event is handled, or through the errFn when there is an
	//error handling it.
	//An event is considered handled when it and all its child events (derived from the
	//Clone function) have called the Ack function.
	//An event is considered to have an error if it or any of its child events (derived from
	//the Clone function) has called the Nack function.
	//An event can also error out if it does not receive all the acknowledgements before
	//the context timeout/cancellation.
	NewEventWithAck(ctx context.Context, payload interface{}, handledFn func(), errFn func(error)) (Event, error)
}
