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

	//Acknowledge that the event is done and further actions on
	//the event are no longer possible.
	//Proposal 2: We do not expose Ack to the plugin developers.
	//            Instead, we do the ack for them (see below)
	Ack()

	//Clone the event with a new context. (deep-copy payload?)
	//Clone fails and return an error if the event is done
	Clone(ctx context.Context) (Event, error)
}

type NewEventerer interface {
	NewEvent(payload interface{}) (Event, error)
	NewEventWithContext(ctx context.Context, payload interface{}) (Event, error)
}
