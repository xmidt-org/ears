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

package mapping

import (
	"encoding/json"
	"fmt"
	"github.com/boriwo/deepcopy"
	"github.com/rs/zerolog/log"
	"github.com/xmidt-org/ears/internal/pkg/syncer"
	"github.com/xmidt-org/ears/pkg/event"
	"github.com/xmidt-org/ears/pkg/filter"
	"github.com/xmidt-org/ears/pkg/hasher"
	"github.com/xmidt-org/ears/pkg/secret"
	"github.com/xmidt-org/ears/pkg/tenant"
	"go.opentelemetry.io/otel/trace"
	"reflect"
	"strings"
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

func (f *Filter) Filter(evt event.Event) []event.Event {
	if f == nil {
		evt.Nack(&filter.InvalidConfigError{
			Err: fmt.Errorf("<nil> pointer filter"),
		})
		return nil
	}
	log.Ctx(evt.Context()).Debug().Str("op", "filter").Str("filterType", "mapping").Str("name", f.Name()).Msg("mapping")
	err := evt.DeepCopy()
	if err != nil {
		log.Ctx(evt.Context()).Error().Str("op", "filter").Str("filterType", "mapping").Str("name", f.Name()).Msg(err.Error())
		if span := trace.SpanFromContext(evt.Context()); span != nil {
			span.AddEvent(err.Error())
		}
		evt.Ack()
		f.logError()
		return []event.Event{}
	}
	isArray := false
	elem, _, _ := evt.GetPathValue(f.config.ArrayPath)
	iterator := []interface{}{elem}
	switch elem := elem.(type) {
	case []interface{}:
		iterator = elem
		isArray = true
	}
	for idx, aElem := range iterator {
		currEvent := evt
		if isArray {
			var err error
			currEvent, err = event.New(evt.Context(), aElem, event.WithMetadata(evt.Metadata()))
			if err != nil {
				log.Ctx(evt.Context()).Error().Str("op", "filter").Str("filterType", "mapping").Str("name", f.Name()).Msg(err.Error())
				if span := trace.SpanFromContext(evt.Context()); span != nil {
					span.AddEvent(err.Error())
				}
				evt.Ack()
				f.logError()
				return []event.Event{}
			}
		}
		obj, _, _ := currEvent.GetPathValue(f.config.Path)
		mapped := false
		for _, m := range f.config.Map {
			from := m.From
			to := m.To
			switch fromStr := from.(type) {
			case string:
				{
					if strings.HasSuffix(fromStr, "}") && strings.HasPrefix(fromStr, "{") {
						from, _, _ = currEvent.GetPathValue(fromStr[1 : len(fromStr)-1])
					}
				}
			}
			switch toStr := to.(type) {
			case string:
				{
					if strings.HasSuffix(toStr, "}") && strings.HasPrefix(toStr, "{") {
						to, _, _ = currEvent.GetPathValue(toStr[1 : len(toStr)-1])
					}
				}
			}
			if reflect.DeepEqual(obj, from) && f.compare(currEvent, m.Comparison) {
				if isArray {
					currEvent.SetPathValue(f.config.Path, deepcopy.DeepCopy(to), true)
					evt.SetPathValue(f.config.ArrayPath+fmt.Sprintf("[%d]", idx), currEvent.Payload(), true)
				} else {
					evt.SetPathValue(f.config.Path, deepcopy.DeepCopy(to), true)
				}
				mapped = true
			}
		}
		if !mapped && f.config.DefaultValue != nil {
			defVal := f.config.DefaultValue
			switch defStr := f.config.DefaultValue.(type) {
			case string:
				{
					if strings.HasSuffix(defStr, "}") && strings.HasPrefix(defStr, "{") {
						defVal, _, _ = currEvent.GetPathValue(defStr[1 : len(defStr)-1])
					}
				}
			}
			if isArray {
				currEvent.SetPathValue(f.config.Path, deepcopy.DeepCopy(defVal), true)
				evt.SetPathValue(f.config.ArrayPath+fmt.Sprintf("[%d]", idx), currEvent.Payload(), true)
			} else {
				evt.SetPathValue(f.config.Path, deepcopy.DeepCopy(defVal), true)
			}
		}
	}
	f.logSuccess()
	return []event.Event{evt}
}

func (f *Filter) compare(evt event.Event, cmp *Comparison) bool {
	if evt == nil || cmp == nil {
		return true
	}
	for _, eq := range cmp.Equal {
		for b, a := range eq {
			var aObj, bObj interface{}
			aObj = a
			bObj = b
			switch aT := a.(type) {
			case string:
				if strings.HasPrefix(aT, "{") && strings.HasSuffix(aT, "}") {
					aObj, _, _ = evt.GetPathValue(aT[1 : len(aT)-1])
				}
			}
			if strings.HasPrefix(b, "{") && strings.HasSuffix(b, "}") {
				bObj, _, _ = evt.GetPathValue(b[1 : len(b)-1])
			}
			if !reflect.DeepEqual(aObj, bObj) {
				return false
			}
		}
	}
	for _, neq := range cmp.NotEqual {
		for b, a := range neq {
			var aObj, bObj interface{}
			aObj = a
			bObj = b
			switch aT := a.(type) {
			case string:
				if strings.HasPrefix(aT, "{") && strings.HasSuffix(aT, "}") {
					aObj, _, _ = evt.GetPathValue(aT[1 : len(aT)-1])
				}
			}
			if strings.HasPrefix(b, "{") && strings.HasSuffix(b, "}") {
				bObj, _, _ = evt.GetPathValue(b[1 : len(b)-1])
			}
			if reflect.DeepEqual(aObj, bObj) {
				return false
			}
		}
	}
	return true
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
		0,
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
