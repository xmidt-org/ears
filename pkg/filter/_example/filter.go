// SPDX-FileCopyrightText: 2025 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package _example

import (
	"github.com/xmidt-org/ears/pkg/tenant"

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
func (f *Filter) Filter(evt event.Event) []event.Event {
	return []event.Event{evt}
}

func (f *Filter) Config() interface{} {
	return nil
}

func (f *Filter) Name() string {
	return ""
}

func (f *Filter) Plugin() string {
	return ""
}

func (f *Filter) Tenant() tenant.Id {
	return tenant.Id{}
}
