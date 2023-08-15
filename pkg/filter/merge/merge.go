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

package merge

import (
	"fmt"
	"github.com/rs/zerolog/log"
	"github.com/xmidt-org/ears/pkg/event"
	"github.com/xmidt-org/ears/pkg/filter"
	"github.com/xmidt-org/ears/pkg/secret"
	"github.com/xmidt-org/ears/pkg/tenant"
	"go.opentelemetry.io/otel/trace"
	"time"
)

func NewFilter(tid tenant.Id, plugin string, name string, config interface{}, secrets secret.Vault) (*Filter, error) {
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
		config:     *cfg,
		name:       name,
		plugin:     plugin,
		tid:        tid,
		currentSec: time.Now().Unix(),
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
	fromObj, _, _ := evt.GetPathValue(f.config.FromPath)
	toObj, _, _ := evt.GetPathValue(f.config.ToPath)
	mergeObj := merge(fromObj, toObj)
	err := evt.DeepCopy()
	if err != nil {
		log.Ctx(evt.Context()).Error().Str("op", "filter").Str("filterType", "merge").Str("name", f.Name()).Msg(err.Error())
		if span := trace.SpanFromContext(evt.Context()); span != nil {
			span.AddEvent(err.Error())
		}
		evt.Ack()
		f.logError()
		return []event.Event{}
	}
	_, _, err = evt.SetPathValue(f.config.ToPath, mergeObj, true)
	if err != nil {
		log.Ctx(evt.Context()).Error().Str("op", "filter").Str("filterType", "merge").Str("name", f.Name()).Msg(err.Error())
		if span := trace.SpanFromContext(evt.Context()); span != nil {
			span.AddEvent(err.Error())
		}
		evt.Ack()
		f.logError()
		return []event.Event{}
	}
	log.Ctx(evt.Context()).Debug().Str("op", "filter").Str("filterType", "merge").Str("name", f.Name()).Msg("merge")
	f.logSuccess()
	return []event.Event{evt}
}

// merge is a helper function to merge maps a and b. b is merged into a and finally a is returned as the result.
func merge(a interface{}, b interface{}) interface{} {
	// we deep merge maps and shallow merge arrays - maybe arrays should be treated as sets and be deep-merged as well
	if a == nil {
		return b
	}
	if b == nil {
		return a
	}
	switch a.(type) {
	case map[string]interface{}:
		switch b.(type) {
		case map[string]interface{}:
			bc := make(map[string]interface{})
			for k, v := range b.(map[string]interface{}) {
				bc[k] = v
			}
			for k, va := range a.(map[string]interface{}) {
				vb := bc[k]
				bc[k] = merge(va, vb)
			}
			return bc
		case []interface{}:
			return b
		default:
			return b
		}
	case []interface{}:
		switch b.(type) {
		case map[string]interface{}:
			return b
		case []interface{}:
			res := make([]interface{}, 0)
			res = append(res, a.([]interface{})...)
			res = append(res, b.([]interface{})...)
			return res
		default:
			return b
		}
	default:
		switch b.(type) {
		case map[string]interface{}:
			return b
		case []interface{}:
			return b
		default:
			return b
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

func (f *Filter) EventSuccessCount() int {
	return f.successCounter
}

func (f *Filter) EventSuccessVelocity() int {
	return f.successVelocityCounter
}

func (f *Filter) EventFilterCount() int {
	return f.filterCounter
}

func (f *Filter) EventFilterVelocity() int {
	return f.filterVelocityCounter
}

func (f *Filter) EventErrorCount() int {
	return f.errorCounter
}

func (f *Filter) EventErrorVelocity() int {
	return f.errorVelocityCounter
}

func (f *Filter) EventTs() int64 {
	return f.currentSec
}
