// SPDX-FileCopyrightText: 2025 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package receiver

import (
	"context"

	"github.com/xmidt-org/ears/internal/pkg/syncer"

	"github.com/xmidt-org/ears/pkg/secret"
	"github.com/xmidt-org/ears/pkg/tenant"

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
	NewReceiver(tid tenant.Id, plugin string, name string, config interface{}, secrets secret.Vault, tableSyncer syncer.DeltaSyncer) (Receiver, error)
}

// NextFn defines the signature of a function that can take in an
// event and process it
type NextFn func(e event.Event)

// Receiver is a plugin that will receive messages and will
// send them to `NextFn`
type Receiver interface {
	Receive(next NextFn) error
	// StopReceiving will stop the receiver from receiving events.
	// This will cause Receive to return.
	StopReceiving(ctx context.Context) error
	// Send a test event into the route
	Trigger(e event.Event)
	//
	Config() interface{}
	Name() string
	Plugin() string
	Tenant() tenant.Id
	EventSuccessCount() int
	EventSuccessVelocity() int
	EventErrorCount() int
	EventErrorVelocity() int
	EventTs() int64
	LogSuccess()
}
