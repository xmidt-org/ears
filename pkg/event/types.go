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
	"github.com/xmidt-org/ears/pkg/tenant"
	"time"
)

//go:generate rm -f testing_mock.go
//go:generate moq -out testing_mock.go . Event

// prefixes for json path expressions

const (
	PAYLOAD   = "payload"
	METADATA  = "metadata"
	TRACE     = "trace"
	TENANT    = "tenant"
	TIMESTAMP = "timestamp"
)

type Event interface {
	//Get the event payload
	Payload() interface{}

	//Get event id
	Id() string

	//Get the event metadata
	Metadata() map[string]interface{}

	//Get the event tenant
	Tenant() tenant.Id

	//Get the event context
	Context() context.Context

	//Get the event creation time
	Created() time.Time

	//Get response string for synchronous routes
	Response() string

	//Set response string for synchronous routes
	SetResponse(response string)

	//Set the event payload
	//Will return an error if the event is done
	SetPayload(payload interface{}) error

	//Set the event metadata
	//Will return an error if the event is done
	SetMetadata(metadata map[string]interface{}) error

	// SetPathValue sets object value at path in either payload or metadata and returns its parent
	// object and parent key if those exist
	SetPathValue(path string, val interface{}, createPath bool) (interface{}, string, error)

	// GetPathValue finds object at path in either payload or metadata and returns such object
	// if one exists along with its parent object and parent key if those exist
	GetPathValue(path string) (interface{}, interface{}, string)

	//Evaluate processes expressions like "my {payload.some.feature} thing", if the expression is a
	// simple path like "{payload.some.feature}" it will act like GetPathValue()
	Evaluate(expression interface{}) (interface{}, interface{}, string)

	//Replace the current event context
	//Will return an error if the event is done
	SetContext(ctx context.Context) error

	//Acknowledge that the event is handled successfully
	//Further actions on the event are no longer possible
	Ack()

	//Acknowledge that there is an error handling the event
	//Further actions on the event are no longer possible
	Nack(err error)

	//Create a child event with payload shallow-copied
	//Clone fails and return an error if the event is already acknowledged
	Clone(ctx context.Context) (Event, error)

	//Deep copy event payload and metadata
	DeepCopy() error

	//TraceId from upstream
	UserTraceId() string
}

type NoAckHandlersError struct {
}

func (e *NoAckHandlersError) Error() string {
	return errs.String("NoAckHandlersError", nil, nil)
}
