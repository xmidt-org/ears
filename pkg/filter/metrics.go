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
	"sync"
	"time"
)

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
	Hash                          HashFct
}

type HashFct func() string

func NewMetricFilter(tableSyncer syncer.DeltaSyncer, hashFct HashFct) MetricFilter {
	return MetricFilter{
		currentSec:         time.Now().Unix(),
		lastMetricWriteSec: time.Now().Unix(),
		tableSyncer:        tableSyncer,
		Hash:               hashFct,
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
		if f.tableSyncer != nil {
			f.tableSyncer.WriteMetrics(hash, f.getLocalMetric())
		}
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
		if f.tableSyncer != nil {
			f.tableSyncer.WriteMetrics(hash, f.getLocalMetric())
		}
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
		if f.tableSyncer != nil {
			f.tableSyncer.WriteMetrics(hash, f.getLocalMetric())
		}
	} else {
		f.Unlock()
	}
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
		"",
	}
	return metrics
}

func (f *MetricFilter) EventSuccessCount() int {
	if f.tableSyncer == nil {
		return f.getLocalMetric().SuccessCount
	} else {
		hash := f.Hash()
		f.tableSyncer.WriteMetrics(hash, f.getLocalMetric())
		return f.tableSyncer.ReadMetrics(hash).SuccessCount
	}
}

func (f *MetricFilter) EventSuccessVelocity() int {
	if f.tableSyncer == nil {
		return f.getLocalMetric().SuccessVelocity
	} else {
		hash := f.Hash()
		f.tableSyncer.WriteMetrics(hash, f.getLocalMetric())
		return f.tableSyncer.ReadMetrics(hash).SuccessVelocity
	}
}

func (f *MetricFilter) EventFilterCount() int {
	if f.tableSyncer == nil {
		return f.getLocalMetric().FilterCount
	} else {
		hash := f.Hash()
		f.tableSyncer.WriteMetrics(hash, f.getLocalMetric())
		return f.tableSyncer.ReadMetrics(hash).FilterCount
	}
}

func (f *MetricFilter) EventFilterVelocity() int {
	if f.tableSyncer == nil {
		return f.getLocalMetric().FilterVelocity
	} else {
		hash := f.Hash()
		f.tableSyncer.WriteMetrics(hash, f.getLocalMetric())
		return f.tableSyncer.ReadMetrics(hash).FilterVelocity
	}
}

func (f *MetricFilter) EventErrorCount() int {
	if f.tableSyncer == nil {
		return f.getLocalMetric().ErrorCount
	} else {
		hash := f.Hash()
		f.tableSyncer.WriteMetrics(hash, f.getLocalMetric())
		return f.tableSyncer.ReadMetrics(hash).ErrorCount
	}
}

func (f *MetricFilter) EventErrorVelocity() int {
	if f.tableSyncer == nil {
		return f.getLocalMetric().ErrorVelocity
	} else {
		hash := f.Hash()
		f.tableSyncer.WriteMetrics(hash, f.getLocalMetric())
		return f.tableSyncer.ReadMetrics(hash).ErrorVelocity
	}
}

func (f *MetricFilter) EventTs() int64 {
	if f.tableSyncer == nil {
		return f.getLocalMetric().LastEventTs
	} else {
		hash := f.Hash()
		f.tableSyncer.WriteMetrics(hash, f.getLocalMetric())
		return f.tableSyncer.ReadMetrics(hash).LastEventTs
	}
}
