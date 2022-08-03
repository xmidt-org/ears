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

package hash

import (
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/rs/zerolog/log"
	"github.com/xmidt-org/ears/pkg/event"
	"github.com/xmidt-org/ears/pkg/filter"
	"github.com/xmidt-org/ears/pkg/secret"
	"github.com/xmidt-org/ears/pkg/tenant"
	"go.opentelemetry.io/otel/trace"
	"hash/fnv"
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
		log.Ctx(evt.Context()).Error().Str("op", "filter").Str("filterType", "hash").Str("name", f.Name()).Msg("cannot hash nil object at " + f.config.FromPath)
		if span := trace.SpanFromContext(evt.Context()); span != nil {
			span.AddEvent("cannot hash nil object at " + f.config.FromPath)
		}
		evt.Ack()
		return []event.Event{}
	}
	var buf []byte
	var err error
	switch obj := obj.(type) {
	case string:
		buf = []byte(obj)
	case []byte:
		buf = obj
	default:
		buf, err = json.Marshal(obj)
		if err != nil {
			log.Ctx(evt.Context()).Error().Str("op", "filter").Str("filterType", "hash").Str("name", f.Name()).Msg(err.Error())
			if span := trace.SpanFromContext(evt.Context()); span != nil {
				span.AddEvent(err.Error())
			}
			evt.Ack()
			return []event.Event{}
		}
	}
	var output interface{}
	switch f.config.HashAlgorithm {
	case "fnv":
		h := fnv.New32a()
		h.Write(buf)
		fnvHash := int(h.Sum32())
		output = fnvHash
	case "md5":
		md5Hash := md5.Sum(buf)
		output = md5Hash[:]
	case "sha1":
		sha1Hash := sha1.Sum(buf)
		output = sha1Hash[:]
	case "sha256":
		sha256Hash := sha256.Sum256(buf)
		output = sha256Hash[:]
	case "hmac-md5":
		if f.config.Key == "" {
			log.Ctx(evt.Context()).Error().Str("op", "filter").Str("filterType", "hash").Str("name", f.Name()).Msg("key required for hmac")
			if span := trace.SpanFromContext(evt.Context()); span != nil {
				span.AddEvent("key required for hmac")
			}
			evt.Ack()
			return []event.Event{}
		}
		h := hmac.New(md5.New, []byte(f.config.Key))
		h.Write(buf)
		md5Hash := h.Sum(nil)
		output = md5Hash[:]
	case "hmac-sha1":
		if f.config.Key == "" {
			log.Ctx(evt.Context()).Error().Str("op", "filter").Str("filterType", "hash").Str("name", f.Name()).Msg("key required for hmac")
			if span := trace.SpanFromContext(evt.Context()); span != nil {
				span.AddEvent("key required for hmac")
			}
			evt.Ack()
			return []event.Event{}
		}
		h := hmac.New(sha1.New, []byte(f.config.Key))
		h.Write(buf)
		sha1Hash := h.Sum(nil)
		output = sha1Hash[:]
	case "hmac-sha256":
		if f.config.Key == "" {
			log.Ctx(evt.Context()).Error().Str("op", "filter").Str("filterType", "hash").Str("name", f.Name()).Msg("key required for hmac")
			if span := trace.SpanFromContext(evt.Context()); span != nil {
				span.AddEvent("key required for hmac")
			}
			evt.Ack()
			return []event.Event{}
		}
		h := hmac.New(sha256.New, []byte(f.config.Key))
		h.Write(buf)
		sha256Hash := h.Sum(nil)
		output = sha256Hash[:]
	default:
		log.Ctx(evt.Context()).Error().Str("op", "filter").Str("filterType", "hash").Str("name", f.Name()).Msg("unsupported hashing algorithm " + f.config.HashAlgorithm)
		if span := trace.SpanFromContext(evt.Context()); span != nil {
			span.AddEvent("unsupported hashing algorithm " + f.config.HashAlgorithm)
		}
		evt.Ack()
		return []event.Event{}
	}
	if f.config.Encoding == "base64" {
		output = base64.StdEncoding.EncodeToString(output.([]byte))
	} else if f.config.Encoding == "hex" {
		output = hex.EncodeToString(output.([]byte))
	}
	path := f.config.FromPath
	if f.config.ToPath != "" {
		path = f.config.ToPath
	}
	err = evt.DeepCopy()
	if err != nil {
		log.Ctx(evt.Context()).Error().Str("op", "filter").Str("filterType", "hash").Str("name", f.Name()).Msg(err.Error())
		if span := trace.SpanFromContext(evt.Context()); span != nil {
			span.AddEvent(err.Error())
		}
		evt.Ack()
		return []event.Event{}
	}
	evt.SetPathValue(path, output, true)
	log.Ctx(evt.Context()).Debug().Str("op", "filter").Str("filterType", "hash").Str("name", f.Name()).Msg("hash")
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
