// SPDX-FileCopyrightText: 2025 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package batch

import (
	"encoding/json"
	"fmt"

	"github.com/xmidt-org/ears/internal/pkg/syncer"
	"github.com/xmidt-org/ears/pkg/hasher"

	//"github.com/gohobby/deepcopy"
	"github.com/boriwo/deepcopy"
	"github.com/rs/zerolog/log"
	"github.com/xmidt-org/ears/pkg/event"
	"github.com/xmidt-org/ears/pkg/filter"
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
	f.batch = make([]event.Event, 0)
	return f, nil
}

func (f *Filter) Filter(evt event.Event) []event.Event {
	if f == nil {
		evt.Nack(&filter.InvalidConfigError{
			Err: fmt.Errorf("<nil> pointer filter"),
		})
		return nil
	}
	f.Lock()
	f.batch = append(f.batch, evt)
	f.Unlock()
	if len(f.batch) >= *f.config.BatchSize {
		newEvt, err := evt.Clone(evt.Context())
		if err != nil {
			log.Ctx(evt.Context()).Error().Str("op", "filter").Str("filterType", "batch").Str("name", f.Name()).Msg(err.Error())
			if span := trace.SpanFromContext(evt.Context()); span != nil {
				span.AddEvent(err.Error())
			}
			newEvt.Ack()
			for _, e := range f.batch {
				f.LogError()
				e.Ack()
			}
			f.Lock()
			f.batch = make([]event.Event, 0)
			f.Unlock()
			return []event.Event{}
		}
		batchPayload := make([]interface{}, 0)
		f.Lock()
		for _, e := range f.batch {
			batchPayload = append(batchPayload, e.Payload())
		}
		f.Unlock()
		err = newEvt.SetMetadata(deepcopy.DeepCopy(evt.Metadata()).(map[string]interface{}))
		if err != nil {
			log.Ctx(evt.Context()).Error().Str("op", "filter").Str("filterType", "batch").Str("name", f.Name()).Msg(err.Error())
			if span := trace.SpanFromContext(evt.Context()); span != nil {
				span.AddEvent(err.Error())
			}
			newEvt.Ack()
			for _, e := range f.batch {
				f.LogError()
				e.Ack()
			}
			f.Lock()
			f.batch = make([]event.Event, 0)
			f.Unlock()
			return []event.Event{}
		}
		err = newEvt.SetPayload(batchPayload)
		if err != nil {
			log.Ctx(evt.Context()).Error().Str("op", "filter").Str("filterType", "batch").Str("name", f.Name()).Msg(err.Error())
			if span := trace.SpanFromContext(evt.Context()); span != nil {
				span.AddEvent(err.Error())
			}
			newEvt.Ack()
			for _, e := range f.batch {
				f.LogError()
				e.Ack()
			}
			f.Lock()
			f.batch = make([]event.Event, 0)
			f.Unlock()
			return []event.Event{}
		}
		for _, e := range f.batch {
			f.LogSuccess()
			e.Ack()
		}
		log.Ctx(evt.Context()).Debug().Str("op", "filter").Str("filterType", "match").Str("name", f.Name()).Int("eventCount", len(f.batch)).Msg("match")
		f.Lock()
		f.batch = make([]event.Event, 0)
		f.Unlock()
		return []event.Event{newEvt}
	}
	log.Ctx(evt.Context()).Debug().Str("op", "filter").Str("filterType", "batch").Str("name", f.Name()).Int("eventCount", 0).Msg("batch")
	return []event.Event{}
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
