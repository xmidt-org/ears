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
	"github.com/xmidt-org/ears/internal/pkg/syncer"
	"github.com/xmidt-org/ears/pkg/secret"
	"github.com/xmidt-org/ears/pkg/tenant"
	"sync"
	"time"

	"github.com/xmidt-org/ears/pkg/event"
)

//go:generate rm -f testing_mock.go
//go:generate moq -out testing_mock.go . Hasher NewFilterer Filterer Chainer

// InvalidConfigError is returned when a configuration parameter
// results in a plugin error
type InvalidConfigError struct {
	Err error
}

type InvalidArgumentError struct {
	Err error
}

// Hasher defines the hashing interface that a receiver
// needs to implement
type Hasher interface {
	// FiltererHash calculates the hash of a filterer based on the
	// given configuration
	FiltererHash(config interface{}) (string, error)
}

// NewFilterer defines the interface on how to
// to create a new filterer
type NewFilterer interface {
	Hasher

	// NewFilterer returns an object that implements the Filterer interface
	NewFilterer(tid tenant.Id, plugin string, name string, config interface{}, secrets secret.Vault, tableSyncer syncer.DeltaSyncer) (Filterer, error)
}

// Filterer defines the interface that a filterer must implement
type Filterer interface {
	Filter(e event.Event) []event.Event
	Config() interface{}
	Name() string
	Plugin() string
	Tenant() tenant.Id
	EventSuccessCount() int
	EventSuccessVelocity() int
	EventFilterCount() int
	EventFilterVelocity() int
	EventErrorCount() int
	EventErrorVelocity() int
	EventTs() int64
}

// Chainer
type Chainer interface {
	Filterer

	// Add will add a filterer to the chain
	Add(f Filterer) error
	Filterers() []Filterer
}

var _ Chainer = (*Chain)(nil)

type Chain struct {
	sync.RWMutex

	filterers []Filterer
}

type MetricFilter struct {
	sync.Mutex
	successCounter                int
	errorCounter                  int
	filterCounter                 int
	successVelocityCounter        int
	errorVelocityCounter          int
	filterVelocityCounter         int
	currentSuccessVelocityCounter int
	currentErrorVelocityCounter   int
	currentFilterVelocityCounter  int
	currentSec                    int64
	lastMetricWriteSec            int64
	tableSyncer                   syncer.DeltaSyncer
}

func NewMetricFilter(tableSyncer syncer.DeltaSyncer) MetricFilter {
	return MetricFilter{
		currentSec:         time.Now().Unix(),
		lastMetricWriteSec: time.Now().Unix(),
		tableSyncer:        tableSyncer,
	}
}

func (f *MetricFilter) LogSuccess() {
	f.Lock()
	f.successCounter++
	if time.Now().Unix() != f.currentSec {
		f.successVelocityCounter = f.currentSuccessVelocityCounter
		f.currentSuccessVelocityCounter = 0
		f.currentSec = time.Now().Unix()
	}
	f.currentSuccessVelocityCounter++
	if time.Now().Unix()-f.lastMetricWriteSec > 60 {
		hash := f.Hash()
		f.lastMetricWriteSec = time.Now().Unix()
		f.Unlock()
		f.tableSyncer.WriteMetrics(hash, f.getLocalMetric())
	} else {
		f.Unlock()
	}
}

func (f *MetricFilter) LogError() {
	f.Lock()
	f.errorCounter++
	if time.Now().Unix() != f.currentSec {
		f.errorVelocityCounter = f.currentErrorVelocityCounter
		f.currentErrorVelocityCounter = 0
		f.currentSec = time.Now().Unix()
	}
	f.currentErrorVelocityCounter++
	if time.Now().Unix()-f.lastMetricWriteSec > 60 {
		hash := f.Hash()
		f.lastMetricWriteSec = time.Now().Unix()
		f.Unlock()
		f.tableSyncer.WriteMetrics(hash, f.getLocalMetric())
	} else {
		f.Unlock()
	}
}

func (f *MetricFilter) LogFilter() {
	f.Lock()
	f.filterCounter++
	if time.Now().Unix() != f.currentSec {
		f.filterVelocityCounter = f.currentFilterVelocityCounter
		f.currentFilterVelocityCounter = 0
		f.currentSec = time.Now().Unix()
	}
	f.currentFilterVelocityCounter++
	if time.Now().Unix()-f.lastMetricWriteSec > 60 {
		hash := f.Hash()
		f.lastMetricWriteSec = time.Now().Unix()
		f.Unlock()
		f.tableSyncer.WriteMetrics(hash, f.getLocalMetric())
	} else {
		f.Unlock()
	}
}

func (f *MetricFilter) Hash() string {
	return ""
}

func (f *MetricFilter) getLocalMetric() *syncer.EarsMetric {
	f.Lock()
	defer f.Unlock()
	metrics := &syncer.EarsMetric{
		f.successCounter,
		f.errorCounter,
		f.filterCounter,
		f.successVelocityCounter,
		f.errorVelocityCounter,
		f.filterVelocityCounter,
		f.currentSec,
		0,
	}
	return metrics
}

func (f *MetricFilter) EventSuccessCount() int {
	hash := f.Hash()
	f.tableSyncer.WriteMetrics(hash, f.getLocalMetric())
	return f.tableSyncer.ReadMetrics(hash).SuccessCount
}

func (f *MetricFilter) EventSuccessVelocity() int {
	hash := f.Hash()
	f.tableSyncer.WriteMetrics(hash, f.getLocalMetric())
	return f.tableSyncer.ReadMetrics(hash).SuccessVelocity
}

func (f *MetricFilter) EventFilterCount() int {
	hash := f.Hash()
	f.tableSyncer.WriteMetrics(hash, f.getLocalMetric())
	return f.tableSyncer.ReadMetrics(hash).FilterCount
}

func (f *MetricFilter) EventFilterVelocity() int {
	hash := f.Hash()
	f.tableSyncer.WriteMetrics(hash, f.getLocalMetric())
	return f.tableSyncer.ReadMetrics(hash).FilterVelocity
}

func (f *MetricFilter) EventErrorCount() int {
	hash := f.Hash()
	f.tableSyncer.WriteMetrics(hash, f.getLocalMetric())
	return f.tableSyncer.ReadMetrics(hash).ErrorCount
}

func (f *MetricFilter) EventErrorVelocity() int {
	hash := f.Hash()
	f.tableSyncer.WriteMetrics(hash, f.getLocalMetric())
	return f.tableSyncer.ReadMetrics(hash).ErrorVelocity
}

func (f *MetricFilter) EventTs() int64 {
	hash := f.Hash()
	f.tableSyncer.WriteMetrics(hash, f.getLocalMetric())
	return f.tableSyncer.ReadMetrics(hash).LastEventTs
}
