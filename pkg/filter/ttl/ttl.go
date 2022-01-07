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
	"errors"
	"fmt"
	"github.com/rs/zerolog/log"
	"github.com/xmidt-org/ears/pkg/event"
	"github.com/xmidt-org/ears/pkg/filter"
	"github.com/xmidt-org/ears/pkg/secret"
	"github.com/xmidt-org/ears/pkg/tenant"
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
		evt.Nack(errors.New("nil object in " + f.name + " at path " + f.config.Path))
		return []event.Event{}
	}
	evtTs, ok := obj.(float64)
	if !ok {
		evt.Nack(errors.New("not a uint64 at path " + f.config.Path))
		return []event.Event{}
	}
	evtTs = evtTs * float64(*f.config.NanoFactor)
	nowNanos := time.Now().UnixNano()
	if int64(evtTs) > nowNanos {
		evt.Nack(errors.New("event from the future in " + f.name + " at path " + f.config.Path))
		return []event.Event{}
	}
	if nowNanos-int64(evtTs) >= int64(*f.config.Ttl*(*f.config.NanoFactor)) {
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
