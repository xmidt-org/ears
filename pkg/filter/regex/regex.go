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

package regex

import (
	"encoding/json"
	"fmt"
	"github.com/rs/zerolog/log"
	"github.com/xmidt-org/ears/pkg/event"
	"github.com/xmidt-org/ears/pkg/filter"
	"github.com/xmidt-org/ears/pkg/secret"
	"github.com/xmidt-org/ears/pkg/tenant"
	"go.opentelemetry.io/otel/trace"
	"regexp"
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
		config: *cfg,
		name:   name,
		plugin: plugin,
		tid:    tid,
	}
	return f, nil
}

func (f *Filter) Filter(evt event.Event) []event.Event {
	if f == nil {
		evt.Nack(&filter.InvalidConfigError{
			Err: fmt.Errorf("<nil> pointer filter"),
		})
		return nil
	}
	obj, _, _ := evt.GetPathValue(f.config.FromPath)
	if obj == nil {
		log.Ctx(evt.Context()).Error().Str("op", "filter").Str("filterType", "regex").Str("name", f.Name()).Msg("nil object at " + f.config.FromPath)
		if span := trace.SpanFromContext(evt.Context()); span != nil {
			span.AddEvent("nil object at " + f.config.FromPath)
		}
		evt.Ack()
		return []event.Event{}
	}
	objAsStr := ""
	switch t := obj.(type) {
	case string:
		objAsStr = t
	case []byte:
		objAsStr = string(t)
	default:
		buf, err := json.Marshal(obj)
		if err != nil {
			log.Ctx(evt.Context()).Error().Str("op", "filter").Str("filterType", "regex").Str("name", f.Name()).Msg(err.Error())
			if span := trace.SpanFromContext(evt.Context()); span != nil {
				span.AddEvent(err.Error())
			}
			evt.Ack()
			return []event.Event{}
		}
		objAsStr = string(buf)
	}
	r, err := regexp.Compile(f.config.Regex)
	if err != nil {
		log.Ctx(evt.Context()).Error().Str("op", "filter").Str("filterType", "regex").Str("name", f.Name()).Msg(err.Error())
		if span := trace.SpanFromContext(evt.Context()); span != nil {
			span.AddEvent(err.Error())
		}
		evt.Ack()
		return []event.Event{}
	}
	var output string
	if f.config.ReplaceAllString != nil {
		output = r.ReplaceAllString(objAsStr, *f.config.ReplaceAllString)
	} else {
		output = r.FindString(objAsStr)
	}
	path := f.config.FromPath
	if f.config.ToPath != "" {
		path = f.config.ToPath
	}
	err = evt.DeepCopy()
	if err != nil {
		log.Ctx(evt.Context()).Error().Str("op", "filter").Str("filterType", "regex").Str("name", f.Name()).Msg(err.Error())
		if span := trace.SpanFromContext(evt.Context()); span != nil {
			span.AddEvent(err.Error())
		}
		evt.Ack()
		return []event.Event{}
	}
	_, _, err = evt.SetPathValue(path, output, true)
	if err != nil {
		log.Ctx(evt.Context()).Error().Str("op", "filter").Str("filterType", "regex").Str("name", f.Name()).Msg(err.Error())
		if span := trace.SpanFromContext(evt.Context()); span != nil {
			span.AddEvent(err.Error())
		}
		evt.Ack()
		return []event.Event{}
	}
	log.Ctx(evt.Context()).Debug().Str("op", "filter").Str("filterType", "regex").Str("name", f.Name()).Msg("regex")
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
