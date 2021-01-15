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

package _example

import (
	"context"

	"github.com/xmidt-org/ears/pkg/event"
	"github.com/xmidt-org/ears/pkg/filter"
)

// NewFilter takes in a configuration and will create a
// new `Filter` structure based upon the configuration.
// NewFilter should validate the configuration and return
// any config validation errors.  `NewConfig` should wrap
// its returned error in an `InvalidConfigErr` struct, so
// wrapping it again is not necessary.
func NewFilter(config interface{}) (*Filter, error) {
	cfg, err := NewConfig(config)
	if err != nil {
		return nil, err
	}

	// Fill in any default values
	cfg = cfg.WithDefaults()

	// Validate the configuration after filling in default values
	err = cfg.Validate()
	if err != nil {
		return nil, err
	}

	// Set up the filter based on the passed in configuration.
	// Then return the instantiated filter.
	return &Filter{}, nil
}

// Compile time check to make sure we have implemented the
// filterer interface properly
var _ filter.Filterer = (*Filter)(nil)

// Filter is the structure that will implement the `Filter()` interface
type Filter struct{}

// Filter the event based on the configuration information.
// This example passes the event on as-is.
func (f *Filter) Filter(ctx context.Context, evt event.Event) ([]event.Event, error) {
	return []event.Event{evt}, nil
}
