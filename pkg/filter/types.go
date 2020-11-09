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

package filter

import (
	"context"

	"github.com/xmidt-org/ears/pkg/event"
)

//go:generate rm testing_mock.go
//go:generate moq -out testing_mock.go . Hasher NewFilterer Filterer

// InvalidConfigError is returned when a configuration parameter
// results in a plugin error
type InvalidConfigError struct {
	Err error
}

// Hasher defines the hashing interface that a receiver
// needs to implement
type Hasher interface {
	// FiltererHash calculates the hash of a filterer based on the
	// given configuration
	FiltererHash(config string) (string, error)
}

// NewFilterer defines the interface on how to
// to create a new filterer
type NewFilterer interface {
	Hasher

	// NewFilterer returns an object that implements the Filterer interface
	NewFilterer(config string) (Filterer, error)
}

// Filterer defines the interface that a filterer must implement
type Filterer interface {
	Filter(ctx context.Context, e event.Event) ([]event.Event, error)
}
