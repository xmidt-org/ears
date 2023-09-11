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

package transform

import (
	"encoding/json"
	"fmt"
	"github.com/xmidt-org/ears/internal/pkg/syncer"
	"github.com/xmidt-org/ears/pkg/hasher"
	"time"

	//"github.com/mohae/deepcopy"
	//"github.com/gohobby/deepcopy"
	"github.com/boriwo/deepcopy"
	"github.com/rs/zerolog/log"
	"github.com/xmidt-org/ears/pkg/event"
	"github.com/xmidt-org/ears/pkg/filter"
	"github.com/xmidt-org/ears/pkg/secret"
	"github.com/xmidt-org/ears/pkg/tenant"
	"go.opentelemetry.io/otel/trace"
	"strings"
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

// Filter splits an event containing an array into multiple events
func (f *Filter) Filter(evt event.Event) []event.Event {
	if f == nil {
		evt.Nack(&filter.InvalidConfigError{
			Err: fmt.Errorf("<nil> pointer filter"),
		})
		return nil
	}
	events := []event.Event{}
	if f.config.Transformation == nil {
		events = append(events, evt)
	} else {
		err := evt.DeepCopy()
		if err != nil {
			log.Ctx(evt.Context()).Error().Str("op", "filter").Str("filterType", "transform").Str("name", f.Name()).Msg(err.Error())
			if span := trace.SpanFromContext(evt.Context()); span != nil {
				span.AddEvent(err.Error())
			}
			evt.Ack()
			f.logError()
			return []event.Event{}
		}
		obj, _, _ := evt.GetPathValue(f.config.FromPath)
		a := []interface{}{obj}
		isArray := false
		switch obj := obj.(type) {
		case []interface{}:
			a = obj
			isArray = true
		}
		for idx, elem := range a {
			subEvt, err := event.New(evt.Context(), elem, event.WithMetadata(evt.Metadata()), event.WithTenant(evt.Tenant()))
			if err != nil {
				log.Ctx(evt.Context()).Error().Str("op", "filter").Str("filterType", "transform").Str("name", f.Name()).Msg(err.Error())
				if span := trace.SpanFromContext(evt.Context()); span != nil {
					span.AddEvent(err.Error())
				}
				evt.Ack()
				f.logError()
				return []event.Event{}
			}
			thisTransform := deepcopy.DeepCopy(f.config.Transformation)
			transform(subEvt, thisTransform, nil, "", -1)
			if isArray {
				_, _, err = evt.SetPathValue(fmt.Sprintf("%s[%d]", f.config.ToPath, idx), thisTransform, true)
			} else {
				_, _, err = evt.SetPathValue(f.config.ToPath, thisTransform, true)
			}
			if err != nil {
				log.Ctx(evt.Context()).Error().Str("op", "filter").Str("filterType", "transform").Str("name", f.Name()).Msg(err.Error())
				if span := trace.SpanFromContext(evt.Context()); span != nil {
					span.AddEvent(err.Error())
				}
				evt.Ack()
				f.logError()
				return []event.Event{}
			}
		}
		events = append(events, evt)
	}
	log.Ctx(evt.Context()).Debug().Str("op", "filter").Str("filterType", "transform").Str("name", f.Name()).Int("eventCount", len(events)).Msg("transform")
	f.logSuccess()
	return events
}

// transform is a helper function to perform a simple transformation
func transform(evt event.Event, t interface{}, parent interface{}, key string, idx int) {
	if evt == nil || t == nil {
		return
	}
	switch tt := t.(type) {
	case map[string]interface{}:
		for key, st := range tt {
			transform(evt, st, tt, key, -1)
		}
	case []interface{}:
		for idx, st := range tt {
			transform(evt, st, tt, "", idx)
		}
	case string:
		for {
			si := strings.Index(tt, "{")
			ei := strings.Index(tt, "}")
			if si < 0 || ei < 0 {
				break
			}
			path := tt[si+1 : ei]
			v, _, _ := evt.GetPathValue(path)
			v = deepcopy.DeepCopy(v)
			if !(si == 0 && ei == len(tt)-1) {
				switch vt := v.(type) {
				case string:
					tt = tt[0:si] + vt + tt[ei+1:]
				default:
					sv, _ := json.Marshal(vt)
					tt = tt[0:si] + string(sv) + tt[ei+1:]
				}
				v = tt
			} else {
				tt = ""
			}
			if parent != nil {
				if key != "" {
					parent.(map[string]interface{})[key] = v
				} else if idx >= 0 {
					parent.([]interface{})[idx] = v
				}
			}
		}
	}
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
