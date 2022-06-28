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

package ttl

import (
	"context"
	"fmt"
	"github.com/rs/zerolog/log"
	"github.com/xmidt-org/ears/internal/pkg/rtsemconv"
	"github.com/xmidt-org/ears/pkg/event"
	"github.com/xmidt-org/ears/pkg/filter"
	"github.com/xmidt-org/ears/pkg/secret"
	"github.com/xmidt-org/ears/pkg/tenant"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/global"
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
		config: *cfg,
		name:   name,
		plugin: plugin,
		tid:    tid,
	}

	// metric recorders
	meter := global.Meter(rtsemconv.EARSMeterName)
	commonLabels := []attribute.KeyValue{
		attribute.String(rtsemconv.EARSPluginTypeLabel, rtsemconv.EARSPluginTypeTtlFilter),
		attribute.String(rtsemconv.EARSPluginNameLabel, f.Name()),
		attribute.String(rtsemconv.EARSAppIdLabel, f.tid.AppId),
		attribute.String(rtsemconv.EARSOrgIdLabel, f.tid.OrgId),
	}
	f.eventTtlExpirationCounter = metric.Must(meter).
		NewInt64Counter(
			rtsemconv.EARSMetricEventTtlExpiration,
			metric.WithDescription("measures the number of expired events"),
		).Bind(commonLabels...)

	return f, nil
}

func (f *Filter) Filter(evt event.Event) []event.Event {
	if f == nil {
		evt.Nack(&filter.InvalidConfigError{
			Err: fmt.Errorf("<nil> pointer filter"),
		})
		return nil
	}
	if *f.config.Disabled {
		return []event.Event{evt}
	}
	log.Ctx(evt.Context()).Debug().Str("op", "filter").Str("filterType", "ttl").Str("name", f.Name()).Msg("ttl")
	obj, _, _ := evt.GetPathValue(f.config.Path)
	if obj == nil {
		log.Ctx(evt.Context()).Error().Str("op", "filter").Str("filterType", "ttl").Str("name", f.Name()).Msg("nil object at " + f.config.Path)
		if span := trace.SpanFromContext(evt.Context()); span != nil {
			span.AddEvent("nil object at " + f.config.Path)
		}
		evt.Ack()
		return []event.Event{}
	}
	evtTs, ok := obj.(float64)
	if !ok {
		log.Ctx(evt.Context()).Error().Str("op", "filter").Str("filterType", "ttl").Str("name", f.Name()).Msg("not a uint64 at path " + f.config.Path)
		if span := trace.SpanFromContext(evt.Context()); span != nil {
			span.AddEvent("not a uint64 at path " + f.config.Path)
		}
		evt.Ack()
		return []event.Event{}
	}
	evtTs = evtTs * float64(*f.config.NanoFactor)
	nowNanos := time.Now().UnixNano()
	if int64(evtTs) > nowNanos {
		log.Ctx(evt.Context()).Error().Str("op", "filter").Str("filterType", "ttl").Str("name", f.Name()).Msg("event from the future at path " + f.config.Path)
		if span := trace.SpanFromContext(evt.Context()); span != nil {
			span.AddEvent("event from the future at path " + f.config.Path)
		}
		evt.Ack()
		return []event.Event{}
	}
	// ttl is in milliseconds, so we need to convert the timestamp nanos
	if (nowNanos-int64(evtTs))/1e6 >= int64(*f.config.Ttl) {
		log.Ctx(evt.Context()).Info().Str("op", "filter").Str("filterType", "ttl").Str("name", f.Name()).Msg("event ttl expired")
		ctx := context.Background()
		f.eventTtlExpirationCounter.Add(ctx, 1)
		evt.Ack()
		return []event.Event{}
	}
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
