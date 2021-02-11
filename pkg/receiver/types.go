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

package receiver

import (
	"context"

	"github.com/xmidt-org/ears/pkg/event"
)

//go:generate rm -f testing_mock.go
//go:generate moq -out testing_mock.go . Hasher NewReceiverer Receiver

// InvalidConfigError is returned when a bad configuration
// is passed into the New* functions
type InvalidConfigError struct {
	Err error
}

// Hasher defines the hashing interface that a receiver
// needs to implement
type Hasher interface {
	// ReceiverHash calculates the hash of a receiver based on the
	// given configuration
	ReceiverHash(config interface{}) (string, error)
}

type NewReceiverer interface {
	Hasher

	// NewReceiver returns an object that implements the
	// Receiver interface
	NewReceiver(config interface{}) (Receiver, error)
}

// NextFn defines the signature of a function that can take in an
// event and process it
type NextFn func(e event.Event) error

// Receiver is a plugin that will receive messages and will
// send them to `NextFn`
type Receiver interface {
	Receive(ctx context.Context, next NextFn) error

	// StopReceiving will stop the receiver from receiving events.
	// This will cause Receive to return.
	StopReceiving(ctx context.Context) error
}
