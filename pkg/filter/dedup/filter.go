// SPDX-FileCopyrightText: 2025 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package dedup

import (
	"crypto/md5"
	"encoding/json"
	"fmt"

	lru "github.com/hashicorp/golang-lru"
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
	f.lruCache, err = lru.New(*cfg.CacheSize)
	if err != nil {
		return nil, err
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
	obj, _, _ := evt.GetPathValue(f.config.Path)
	if obj == nil {
		log.Ctx(evt.Context()).Error().Str("op", "filter").Str("filterType", "dedup").Str("name", f.Name()).Msg("nil object at " + f.config.Path)
		if span := trace.SpanFromContext(evt.Context()); span != nil {
			span.AddEvent("nil object at " + f.config.Path)
		}
		evt.Ack()
		f.LogError()
		return []event.Event{}
	}
	buf, err := json.Marshal(obj)
	if err != nil {
		log.Ctx(evt.Context()).Error().Str("op", "filter").Str("filterType", "dedup").Str("name", f.Name()).Msg(err.Error())
		if span := trace.SpanFromContext(evt.Context()); span != nil {
			span.AddEvent(err.Error())
		}
		evt.Ack()
		f.LogError()
		return []event.Event{}
	}
	evtHash := fmt.Sprintf("%x", md5.Sum(buf))
	_, ok := f.lruCache.Get(evtHash)
	if !ok {
		f.lruCache.Add(evtHash, struct{}{})
		log.Ctx(evt.Context()).Debug().Str("op", "filter").Str("filterType", "dedup").Str("name", f.Name()).Int("eventCount", 1).Msg("dedup")
		f.LogSuccess()
		return []event.Event{evt}
	} else {
		evt.Ack()
		log.Ctx(evt.Context()).Debug().Str("op", "filter").Str("filterType", "dedup").Str("name", f.Name()).Int("eventCount", 0).Msg("dedup")
		f.LogFilter()
		return []event.Event{}
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
