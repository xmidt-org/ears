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
	"github.com/xmidt-org/ears/internal/pkg/ack"

	"github.com/mohae/deepcopy"
)

type event struct {
	metadata interface{}
	payload  interface{}
	ctx      context.Context
	ack      ack.SubTree
}

type EventOption func(*event) error

//Create a new event given a context, a payload, and other event options
func New(ctx context.Context, payload interface{}, options ...EventOption) (Event, error) {
	e := &event{
		payload: payload,
		ctx:     ctx,
		ack:     nil,
	}
	for _, option := range options {
		err := option(e)
		if err != nil {
			return nil, err
		}
	}

	return e, nil
}

//event acknowledge option with two completion functions,
//handledFn and errFn. An event with acknowledgement option will be notified through
//the handledFn when an event is handled, or through the errFn when there is an
//error handling it.
//An event is considered handled when it and all its child events (derived from the
//Clone function) have called the Ack function.
//An event is considered to have an error if it or any of its child events (derived from
//the Clone function) has called the Nack function.
//An event can also error out if it does not receive all the acknowledgements before
//the context timeout/cancellation.
func WithAck(handledFn func(Event), errFn func(Event, error)) EventOption {
	return func(e *event) error {
		if handledFn == nil || errFn == nil {
			return &NoAckHandlersError{}
		}

		e.ack = ack.NewAckTree(e.ctx, func() {
			handledFn(e)
		}, func(err error) {
			errFn(e, err)
		})
		return nil
	}
}

func WithMetadata(metadata interface{}) EventOption {
	return func(e *event) error {
		e.SetMetadata(metadata)
		return nil
	}
}

func (e *event) Payload() interface{} {
	return e.payload
}

func (e *event) SetPayload(payload interface{}) error {
	if e.ack != nil && e.ack.IsAcked() {
		return &ack.AlreadyAckedError{}
	}
	e.payload = payload
	return nil
}

func (e *event) Metadata() interface{} {
	return e.metadata
}

func (e *event) SetMetadata(metadata interface{}) error {
	if e.ack != nil && e.ack.IsAcked() {
		return &ack.AlreadyAckedError{}
	}
	e.metadata = metadata
	return nil
}
func (e *event) Context() context.Context {
	return e.ctx
}

func (e *event) SetContext(ctx context.Context) error {
	if e.ack != nil && e.ack.IsAcked() {
		return &ack.AlreadyAckedError{}
	}
	e.ctx = ctx
	return nil
}

func (e *event) Ack() {
	if e.ack != nil {
		e.ack.Ack()
	}
}

func (e *event) Nack(err error) {
	if e.ack != nil {
		e.ack.Nack(err)
	}
}

func (e *event) Clone(ctx context.Context) (Event, error) {
	var subTree ack.SubTree
	if e.ack != nil {
		var err error
		subTree, err = e.ack.NewSubTree()
		if err != nil {
			return nil, err
		}
	}
	//Very fast according to the following benchmark:
	//https://xuri.me/2018/06/17/deep-copy-object-with-reflecting-or-gob-in-go.html
	//Unclear if all types are supported:
	//https://github.com/mohae/deepcopy/blob/master/deepcopy.go#L45-L46
	newPayloadCopy := deepcopy.Copy(e.payload)
	newMetadtaCopy := deepcopy.Copy(e.metadata)
	return &event{
		payload:  newPayloadCopy,
		metadata: newMetadtaCopy,
		ctx:      ctx,
		ack:      subTree,
	}, nil
}
