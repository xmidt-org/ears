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
	payload interface{}
	ctx     context.Context
	ack     ack.SubTree
}

func NewEvent(ctx context.Context, payload interface{}) (Event, error) {
	return &event{
		payload: payload,
		ctx:     ctx,
		ack:     nil,
	}, nil
}

func NewEventWithAck(ctx context.Context, payload interface{}, handledFn func(), errFn func(error)) (Event, error) {
	return &event{
		payload: payload,
		ctx:     ctx,
		ack:     ack.NewAckTree(ctx, handledFn, errFn),
	}, nil
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
	newCopy := deepcopy.Copy(e.payload)

	return &event{
		payload: newCopy,
		ctx:     ctx,
		ack:     subTree,
	}, nil
}
