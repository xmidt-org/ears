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

package modify

import (
	"fmt"
	"github.com/rs/zerolog/log"
	"github.com/xmidt-org/ears/pkg/event"
	"github.com/xmidt-org/ears/pkg/filter"
	"github.com/xmidt-org/ears/pkg/secret"
	"github.com/xmidt-org/ears/pkg/tenant"
	"strings"
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

func (f *Filter) Filter(evt event.Event) []event.Event {
	if f == nil {
		evt.Nack(&filter.InvalidConfigError{
			Err: fmt.Errorf("<nil> pointer filter"),
		})
		return nil
	}
	log.Ctx(evt.Context()).Debug().Str("op", "filter").Str("filterType", "modify").Str("name", f.Name()).Msg("ttl")
	allPaths := make([]string, 0)
	if f.config.Path != "" {
		allPaths = append(allPaths, f.config.Path)
	}
	if len(f.config.Paths) > 0 {
		allPaths = append(allPaths, f.config.Paths...)
	}
	for _, p := range allPaths {
		obj, _, _ := evt.GetPathValue(p)
		if obj == nil {
			continue
			/*log.Ctx(evt.Context()).Error().Str("op", "filter").Str("filterType", "modify").Str("name", f.Name()).Msg("nil object at " + p)
			if span := trace.SpanFromContext(evt.Context()); span != nil {
				span.AddEvent("nil object at " + p)
			}
			evt.Ack()
			return []event.Event{}*/
		}
		switch strObj := obj.(type) {
		case string:
			if *f.config.ToUpper {
				evt.SetPathValue(p, strings.ToUpper(strObj), false)
			} else if *f.config.ToLower {
				evt.SetPathValue(p, strings.ToLower(strObj), false)
			}
		default:
			continue
			/*log.Ctx(evt.Context()).Error().Str("op", "filter").Str("filterType", "modify").Str("name", f.Name()).Msg("not string type at " + p)
			if span := trace.SpanFromContext(evt.Context()); span != nil {
				span.AddEvent("not string type at " + p)
			}
			evt.Ack()
			return []event.Event{}*/
		}
	}
	f.logSuccess()
	return []event.Event{evt}
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
