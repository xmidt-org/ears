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

package plugin

import (
	"github.com/xmidt-org/ears/internal/pkg/syncer"
	"sync"
	"time"
)

type MetricPlugin struct {
	sync.Mutex
	successCounter                int
	errorCounter                  int
	successVelocityCounter        int
	errorVelocityCounter          int
	currentSuccessVelocityCounter int
	currentErrorVelocityCounter   int
	currentSec                    int64
	lastMetricWriteSec            int64
	tableSyncer                   syncer.DeltaSyncer
}

func NewMetricPlugin(tableSyncer syncer.DeltaSyncer) MetricPlugin {
	return MetricPlugin{
		currentSec:         time.Now().Unix(),
		lastMetricWriteSec: time.Now().Unix(),
		tableSyncer:        tableSyncer,
	}
}

func (f *MetricPlugin) LogSuccess() {
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

func (f *MetricPlugin) LogError() {
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

func (f *MetricPlugin) Hash() string {
	return ""
}

func (f *MetricPlugin) getLocalMetric() *syncer.EarsMetric {
	f.Lock()
	defer f.Unlock()
	metrics := &syncer.EarsMetric{
		f.successCounter,
		f.errorCounter,
		0,
		f.successVelocityCounter,
		f.errorVelocityCounter,
		0,
		f.currentSec,
		0,
	}
	return metrics
}

func (f *MetricPlugin) EventSuccessCount() int {
	hash := f.Hash()
	f.tableSyncer.WriteMetrics(hash, f.getLocalMetric())
	return f.tableSyncer.ReadMetrics(hash).SuccessCount
}

func (f *MetricPlugin) EventSuccessVelocity() int {
	hash := f.Hash()
	f.tableSyncer.WriteMetrics(hash, f.getLocalMetric())
	return f.tableSyncer.ReadMetrics(hash).SuccessVelocity
}

func (f *MetricPlugin) EventErrorCount() int {
	hash := f.Hash()
	f.tableSyncer.WriteMetrics(hash, f.getLocalMetric())
	return f.tableSyncer.ReadMetrics(hash).ErrorCount
}

func (f *MetricPlugin) EventErrorVelocity() int {
	hash := f.Hash()
	f.tableSyncer.WriteMetrics(hash, f.getLocalMetric())
	return f.tableSyncer.ReadMetrics(hash).ErrorVelocity
}

func (f *MetricPlugin) EventTs() int64 {
	hash := f.Hash()
	f.tableSyncer.WriteMetrics(hash, f.getLocalMetric())
	return f.tableSyncer.ReadMetrics(hash).LastEventTs
}
