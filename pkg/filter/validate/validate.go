// SPDX-FileCopyrightText: 2025 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package validate

import (
	"encoding/json"
	"fmt"

	"github.com/rs/zerolog/log"
	"github.com/xeipuuv/gojsonschema"
	"github.com/xmidt-org/ears/internal/pkg/syncer"
	"github.com/xmidt-org/ears/pkg/event"
	"github.com/xmidt-org/ears/pkg/filter"
	"github.com/xmidt-org/ears/pkg/hasher"
	"github.com/xmidt-org/ears/pkg/secret"
	"github.com/xmidt-org/ears/pkg/tenant"
	"go.opentelemetry.io/otel/trace"
)

var _ filter.Filterer = (*Filter)(nil)

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
		schema: gojsonschema.NewGoLoader(cfg.Schema),
	}
	f.MetricFilter = filter.NewMetricFilter(tableSyncer, f.Hash)
	return f, nil
}

// Filter log event and pass it on
func (f *Filter) Filter(evt event.Event) []event.Event {
	if f == nil {
		evt.Nack(&filter.InvalidConfigError{
			Err: fmt.Errorf("<nil> pointer filter"),
		})
		return nil
	}
	obj, _, _ := evt.GetPathValue(f.config.Path)
	doc := gojsonschema.NewGoLoader(obj)
	result, err := gojsonschema.Validate(f.schema, doc)
	if err != nil {
		log.Ctx(evt.Context()).Error().Str("op", "filter").Str("filterType", "validate").Str("name", f.Name()).Msg(err.Error())
		if span := trace.SpanFromContext(evt.Context()); span != nil {
			span.AddEvent(err.Error())
		}
		evt.Ack()
		f.LogError()
		return []event.Event{}
	}
	if !result.Valid() {
		log.Ctx(evt.Context()).Error().Str("op", "filter").Str("filterType", "validate").Str("name", f.Name()).Msg(fmt.Sprintf("%+v", result.Errors()))
		if span := trace.SpanFromContext(evt.Context()); span != nil {
			span.AddEvent(fmt.Sprintf("%+v", result.Errors()))
		}
		evt.Ack()
		f.LogFilter()
		return []event.Event{}
	}
	f.LogSuccess()
	return []event.Event{evt}
}

func (f *Filter) Config() interface{} {
	if f == nil {
		return Config{}
	}
	return nil
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
