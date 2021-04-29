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

import (
	"context"
	"github.com/xmidt-org/ears/pkg/errs"
)

//go:generate rm -f testing_mock.go
//go:generate moq -out testing_mock.go . Event

type Event interface {
	//Get the event payload
	Payload() interface{}

	//Get the event metadata
	Metadata() interface{}

	//Get the event context
	Context() context.Context

	//Set the event payload
	//Will return an error if the event is done
	SetPayload(payload interface{}) error

	//Set the event metadata
	//Will return an error if the event is done
	SetMetadata(metadata interface{}) error

	// Eval finds object at path in either payload or metadata and returns such object
	// if one exists along with its parent object and parent key if those exist
	Eval(path string, metadata bool, create bool) (interface{}, interface{}, string)

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

type NoAckHandlersError struct {
}

func (e *NoAckHandlersError) Error() string {
	return errs.String("NoAckHandlersError", nil, nil)
}
