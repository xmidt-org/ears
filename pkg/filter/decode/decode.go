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

package decode

import (
	"encoding/base64"
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
		config: *cfg,
		name:   name,
		plugin: plugin,
		tid:    tid,
	}
	f.MetricFilter = filter.NewMetricFilter(tableSyncer, f.Hash)
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
		log.Ctx(evt.Context()).Error().Str("op", "filter").Str("filterType", "decode").Str("name", f.Name()).Msg("nil object at " + f.config.FromPath)
		if span := trace.SpanFromContext(evt.Context()); span != nil {
			span.AddEvent("nil object at " + f.config.FromPath)
		}
		evt.Ack()
		f.LogError()
		return []event.Event{}
	}
	var input string
	switch obj := obj.(type) {
	case string:
		input = obj
	case []byte:
		input = string(obj)
	default:
		log.Ctx(evt.Context()).Error().Str("op", "filter").Str("filterType", "decode").Str("name", f.Name()).Msg("unsupported data type at " + f.config.FromPath)
		if span := trace.SpanFromContext(evt.Context()); span != nil {
			span.AddEvent("unsupported data type at " + f.config.FromPath)
		}
		evt.Ack()
		f.LogError()
		return []event.Event{}
	}
	var buf []byte
	var err error
	if f.config.Encoding == "base64" {
		buf, err = base64.StdEncoding.DecodeString(input)
		if err != nil {
			log.Ctx(evt.Context()).Error().Str("op", "filter").Str("filterType", "decode").Str("name", f.Name()).Msg(err.Error())
			if span := trace.SpanFromContext(evt.Context()); span != nil {
				span.AddEvent(err.Error())
			}
			evt.Ack()
			f.LogError()
			return []event.Event{}
		}
	} else {
		buf = []byte(input)
	}
	var output interface{}
	err = json.Unmarshal(buf, &output)
	if err != nil {
		log.Ctx(evt.Context()).Error().Str("op", "filter").Str("filterType", "decode").Str("name", f.Name()).Msg(err.Error())
		if span := trace.SpanFromContext(evt.Context()); span != nil {
			span.AddEvent(err.Error())
		}
		evt.Ack()
		f.LogError()
		return []event.Event{}
	}
	path := f.config.FromPath
	if f.config.ToPath != "" {
		path = f.config.ToPath
	}
	err = evt.DeepCopy()
	if err != nil {
		log.Ctx(evt.Context()).Error().Str("op", "filter").Str("filterType", "decode").Str("name", f.Name()).Msg(err.Error())
		if span := trace.SpanFromContext(evt.Context()); span != nil {
			span.AddEvent(err.Error())
		}
		evt.Ack()
		f.LogError()
		return []event.Event{}
	}
	_, _, err = evt.SetPathValue(path, output, true)
	if err != nil {
		log.Ctx(evt.Context()).Error().Str("op", "filter").Str("filterType", "decode").Str("name", f.Name()).Msg(err.Error())
		if span := trace.SpanFromContext(evt.Context()); span != nil {
			span.AddEvent(err.Error())
		}
		evt.Ack()
		f.LogError()
		return []event.Event{}
	}
	log.Ctx(evt.Context()).Debug().Str("op", "filter").Str("filterType", "decode").Str("name", f.Name()).Msg("decode")
	f.LogSuccess()
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
