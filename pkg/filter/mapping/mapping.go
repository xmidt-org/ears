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

package mapping

import (
	"fmt"
	"github.com/mohae/deepcopy"
	"github.com/rs/zerolog/log"
	"github.com/xmidt-org/ears/pkg/event"
	"github.com/xmidt-org/ears/pkg/filter"
	"github.com/xmidt-org/ears/pkg/secret"
	"github.com/xmidt-org/ears/pkg/tenant"
	"reflect"
	"strings"
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
	log.Ctx(evt.Context()).Debug().Str("op", "filter").Str("filterType", "mapping").Str("name", f.Name()).Msg("mapping")
	for _, path := range f.config.Paths {
		obj, _, _ := evt.GetPathValue(path)
		mapped := false
		for _, m := range f.config.Map {
			from := m.From
			to := m.To
			switch fromStr := from.(type) {
			case string:
				{
					if strings.HasSuffix(fromStr, "}") && strings.HasPrefix(fromStr, "{") {
						from, _, _ = evt.GetPathValue(fromStr[1 : len(fromStr)-1])
					}
				}
			}
			switch toStr := to.(type) {
			case string:
				{
					if strings.HasSuffix(toStr, "}") && strings.HasPrefix(toStr, "{") {
						to, _, _ = evt.GetPathValue(toStr[1 : len(toStr)-1])
					}
				}
			}
			if reflect.DeepEqual(obj, from) && f.compare(evt, m.Comparison) {
				evt.SetPathValue(path, deepcopy.Copy(to), true)
				mapped = true
			}
		}
		if !mapped && f.config.DefaultValue != nil {
			evt.SetPathValue(path, deepcopy.Copy(f.config.DefaultValue), false)
		}
	}
	return []event.Event{evt}
}

func (f *Filter) compare(evt event.Event, cmp *Comparison) bool {
	if evt == nil || cmp == nil {
		return true
	}
	for _, eq := range cmp.Equal {
		for b, a := range eq {
			var aObj, bObj interface{}
			aObj = a
			bObj = b
			switch aT := a.(type) {
			case string:
				if strings.HasPrefix(aT, "{") && strings.HasSuffix(aT, "}") {
					aObj, _, _ = evt.GetPathValue(aT[1 : len(aT)-1])
				}
			}
			if strings.HasPrefix(b, "{") && strings.HasSuffix(b, "}") {
				bObj, _, _ = evt.GetPathValue(b[1 : len(b)-1])
			}
			if !reflect.DeepEqual(aObj, bObj) {
				return false
			}
		}
	}
	for _, neq := range cmp.NotEqual {
		for b, a := range neq {
			var aObj, bObj interface{}
			aObj = a
			bObj = b
			switch aT := a.(type) {
			case string:
				if strings.HasPrefix(aT, "{") && strings.HasSuffix(aT, "}") {
					aObj, _, _ = evt.GetPathValue(aT[1 : len(aT)-1])
				}
			}
			if strings.HasPrefix(b, "{") && strings.HasSuffix(b, "}") {
				bObj, _, _ = evt.GetPathValue(b[1 : len(b)-1])
			}
			if reflect.DeepEqual(aObj, bObj) {
				return false
			}
		}
	}
	return true
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
