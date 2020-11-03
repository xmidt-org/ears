// Copyright 2020 Comcast Cable Communications Management, LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"context"
	"fmt"

	"github.com/xmidt-org/ears/pkg/event"
	pkgreceiver "github.com/xmidt-org/ears/pkg/receiver"
)

var _ pkgreceiver.Receiver = (*receiver)(nil)

type receiver struct {
	nextFn pkgreceiver.NextFn
	config string
}

func (r *receiver) NewReceiver(config string) (pkgreceiver.Receiver, error) {
	return &receiver{
		config: config,
	}, nil
}

func (r *receiver) Receive(ctx context.Context, next pkgreceiver.NextFn) error {
	if next == nil {
		return fmt.Errorf("next cannot be nil")
	}

	r.nextFn = next
	return nil
}

func (r *receiver) Trigger(ctx context.Context, event event.Event) error {
	return r.nextFn(event)
}
