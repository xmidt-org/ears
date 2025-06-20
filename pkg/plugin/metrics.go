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
	Hash                          HashFct
}

type HashFct func() string

func NewMetricPlugin(tableSyncer syncer.DeltaSyncer, hashFct HashFct) MetricPlugin {
	return MetricPlugin{
		currentSec:         time.Now().Unix(),
		lastMetricWriteSec: time.Now().Unix(),
		tableSyncer:        tableSyncer,
		Hash:               hashFct,
	}
}

func (mp *MetricPlugin) LogSuccess() {
	mp.Lock()
	mp.successCounter++
	if time.Now().Unix() != mp.currentSec {
		mp.successVelocityCounter = mp.currentSuccessVelocityCounter
		mp.currentSuccessVelocityCounter = 0
		mp.currentSec = time.Now().Unix()
	}
	mp.currentSuccessVelocityCounter++
	if time.Now().Unix()-mp.lastMetricWriteSec > 60 {
		hash := mp.Hash()
		mp.lastMetricWriteSec = time.Now().Unix()
		mp.Unlock()
		if mp.tableSyncer != nil {
			mp.tableSyncer.WriteMetrics(hash, mp.getLocalMetric())
		}
	} else {
		mp.Unlock()
	}
}

func (mp *MetricPlugin) LogError() {
	mp.Lock()
	mp.errorCounter++
	if time.Now().Unix() != mp.currentSec {
		mp.errorVelocityCounter = mp.currentErrorVelocityCounter
		mp.currentErrorVelocityCounter = 0
		mp.currentSec = time.Now().Unix()
	}
	mp.currentErrorVelocityCounter++
	if time.Now().Unix()-mp.lastMetricWriteSec > 60 {
		hash := mp.Hash()
		mp.lastMetricWriteSec = time.Now().Unix()
		mp.Unlock()
		if mp.tableSyncer != nil {
			mp.tableSyncer.WriteMetrics(hash, mp.getLocalMetric())
		}
	} else {
		mp.Unlock()
	}
}

func (mp *MetricPlugin) getLocalMetric() *syncer.EarsMetric {
	mp.Lock()
	defer mp.Unlock()
	metrics := &syncer.EarsMetric{
		mp.successCounter,
		mp.errorCounter,
		0,
		mp.successVelocityCounter,
		mp.errorVelocityCounter,
		0,
		mp.currentSec,
		0,
		"",
	}
	return metrics
}

func (mp *MetricPlugin) DeleteMetrics() {
	if mp.tableSyncer != nil {
		hash := mp.Hash()
		mp.tableSyncer.DeleteMetrics(hash)
	}
}

func (mp *MetricPlugin) EventSuccessCount() int {
	if mp.tableSyncer == nil {
		return mp.getLocalMetric().SuccessCount
	} else {
		hash := mp.Hash()
		mp.tableSyncer.WriteMetrics(hash, mp.getLocalMetric())
		return mp.tableSyncer.ReadMetrics(hash).SuccessCount
	}
}

func (mp *MetricPlugin) EventSuccessVelocity() int {
	if mp.tableSyncer == nil {
		return mp.getLocalMetric().SuccessVelocity
	} else {
		hash := mp.Hash()
		mp.tableSyncer.WriteMetrics(hash, mp.getLocalMetric())
		return mp.tableSyncer.ReadMetrics(hash).SuccessVelocity
	}
}

func (mp *MetricPlugin) EventErrorCount() int {
	if mp.tableSyncer == nil {
		return mp.getLocalMetric().ErrorCount
	} else {
		hash := mp.Hash()
		mp.tableSyncer.WriteMetrics(hash, mp.getLocalMetric())
		return mp.tableSyncer.ReadMetrics(hash).ErrorCount
	}
}

func (mp *MetricPlugin) EventErrorVelocity() int {
	if mp.tableSyncer == nil {
		return mp.getLocalMetric().ErrorVelocity
	} else {
		hash := mp.Hash()
		mp.tableSyncer.WriteMetrics(hash, mp.getLocalMetric())
		return mp.tableSyncer.ReadMetrics(hash).ErrorVelocity
	}
}

func (mp *MetricPlugin) EventTs() int64 {
	if mp.tableSyncer == nil {
		return mp.getLocalMetric().LastEventTs
	} else {
		hash := mp.Hash()
		mp.tableSyncer.WriteMetrics(hash, mp.getLocalMetric())
		return mp.tableSyncer.ReadMetrics(hash).LastEventTs
	}
}
