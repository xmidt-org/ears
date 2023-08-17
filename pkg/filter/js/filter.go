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

package js

import (
	"encoding/json"
	"fmt"
	"github.com/rs/zerolog/log"
	"github.com/xmidt-org/ears/internal/pkg/syncer"
	"github.com/xmidt-org/ears/pkg/event"
	"github.com/xmidt-org/ears/pkg/filter"
	"github.com/xmidt-org/ears/pkg/hasher"
	"github.com/xmidt-org/ears/pkg/secret"
	"github.com/xmidt-org/ears/pkg/tenant"
	"go.opentelemetry.io/otel/trace"
	"time"
)

func NewFilter(tid tenant.Id, plugin string, name string, config interface{}, secrets secret.Vault, tableSyncer syncer.DeltaSyncer) (*Filter, error) {
	cfg, err := NewConfig(config)
	if err != nil {
		return nil, &filter.InvalidConfigError{
			Err: err,
		}
	}
	cfg = cfg.WithDefaults()
	err = cfg.Validate()
	if err != nil {
		return nil, err
	}
	f := &Filter{
		config:      *cfg,
		name:        name,
		plugin:      plugin,
		tid:         tid,
		currentSec:  time.Now().Unix(),
		tableSyncer: tableSyncer,
	}
	return f, nil
}

func (f *Filter) logSuccess() {
	f.Lock()
	f.successCounter++
	if time.Now().Unix() != f.currentSec {
		f.successVelocityCounter = f.currentSuccessVelocityCounter
		f.currentSuccessVelocityCounter = 0
		f.currentSec = time.Now().Unix()
	}
	f.currentSuccessVelocityCounter++
	f.Unlock()
}

func (f *Filter) logError() {
	f.Lock()
	f.errorCounter++
	if time.Now().Unix() != f.currentSec {
		f.errorVelocityCounter = f.currentErrorVelocityCounter
		f.currentErrorVelocityCounter = 0
		f.currentSec = time.Now().Unix()
	}
	f.currentErrorVelocityCounter++
	f.Unlock()
}

func (f *Filter) logFilter() {
	f.Lock()
	f.filterCounter++
	if time.Now().Unix() != f.currentSec {
		f.filterVelocityCounter = f.currentFilterVelocityCounter
		f.currentFilterVelocityCounter = 0
		f.currentSec = time.Now().Unix()
	}
	f.currentFilterVelocityCounter++
	f.Unlock()
}

// Filter executes javascript transformation
func (f *Filter) Filter(evt event.Event) []event.Event {
	if f == nil {
		evt.Nack(&filter.InvalidConfigError{
			Err: fmt.Errorf("<nil> pointer filter"),
		})
		f.logError()
		return nil
	}
	transformedEvts, err := defaultInterpreter.Exec(evt, f.config.Source)
	if err != nil {
		log.Ctx(evt.Context()).Error().Str("op", "filter").Str("filterType", "js").Str("name", f.Name()).Msg("js filter error: " + err.Error())
		if span := trace.SpanFromContext(evt.Context()); span != nil {
			span.AddEvent("js filter error: " + err.Error())
		}
		evt.Ack()
		f.logError()
		return []event.Event{}
	}
	log.Ctx(evt.Context()).Debug().Str("op", "filter").Str("filterType", "js").Str("name", f.Name()).Int("eventCount", len(transformedEvts)).Msg("js")
	if len(transformedEvts) == 0 {
		evt.Ack()
		f.logFilter()
		return []event.Event{}
	}
	f.logSuccess()
	return transformedEvts
}

func (f *Filter) Config() interface{} {
	if f == nil {
		return Config{}
	}
	return f.config
}

func (f *Filter) Name() string {
	return f.name
}

func (f *Filter) Plugin() string {
	return f.plugin
}

func (f *Filter) Tenant() tenant.Id {
	return f.tid
}

func (f *Filter) getLocalMetric() *syncer.EarsMetric {
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
	}
	return metrics
}

func (f *Filter) EventSuccessCount() int {
	hash := f.Hash()
	f.tableSyncer.WriteMetrics(hash, f.getLocalMetric())
	return f.tableSyncer.ReadMetrics(hash).SuccessCount
}

func (f *Filter) EventSuccessVelocity() int {
	hash := f.Hash()
	f.tableSyncer.WriteMetrics(hash, f.getLocalMetric())
	return f.tableSyncer.ReadMetrics(hash).SuccessVelocity
}

func (f *Filter) EventFilterCount() int {
	hash := f.Hash()
	f.tableSyncer.WriteMetrics(hash, f.getLocalMetric())
	return f.tableSyncer.ReadMetrics(hash).FilterCount
}

func (f *Filter) EventFilterVelocity() int {
	hash := f.Hash()
	f.tableSyncer.WriteMetrics(hash, f.getLocalMetric())
	return f.tableSyncer.ReadMetrics(hash).FilterVelocity
}

func (f *Filter) EventErrorCount() int {
	hash := f.Hash()
	f.tableSyncer.WriteMetrics(hash, f.getLocalMetric())
	return f.tableSyncer.ReadMetrics(hash).ErrorCount
}

func (f *Filter) EventErrorVelocity() int {
	hash := f.Hash()
	f.tableSyncer.WriteMetrics(hash, f.getLocalMetric())
	return f.tableSyncer.ReadMetrics(hash).ErrorVelocity
}

func (f *Filter) EventTs() int64 {
	hash := f.Hash()
	f.tableSyncer.WriteMetrics(hash, f.getLocalMetric())
	return f.tableSyncer.ReadMetrics(hash).LastEventTs
}

func (f *Filter) Hash() string {
	cfg := ""
	if f.Config() != nil {
		buf, _ := json.Marshal(f.Config())
		if buf != nil {
			cfg = string(buf)
		}
	}
	str := f.name + f.plugin + cfg
	hash := hasher.String(str)
	return hash
}
