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

package transform

import (
	"encoding/json"
	"fmt"
	"github.com/mohae/deepcopy"
	"github.com/rs/zerolog/log"
	"github.com/xmidt-org/ears/pkg/event"
	"github.com/xmidt-org/ears/pkg/filter"
	"github.com/xmidt-org/ears/pkg/secret"
	"github.com/xmidt-org/ears/pkg/tenant"
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

// Filter splits an event containing an array into multiple events
func (f *Filter) Filter(evt event.Event) []event.Event {
	if f == nil {
		evt.Nack(&filter.InvalidConfigError{
			Err: fmt.Errorf("<nil> pointer filter"),
		})
		return nil
	}
	events := []event.Event{}
	if f.config.Transformation == nil {
		events = append(events, evt)
	} else {
		obj, _, _ := evt.GetPathValue(f.config.FromPath)
		a := []interface{}{obj}
		isArray := false
		switch obj := obj.(type) {
		case []interface{}:
			a = obj
			isArray = true
		}
		for idx, elem := range a {
			subEvt, err := event.New(evt.Context(), elem, event.WithMetadata(evt.Metadata()), event.WithTenant(evt.Tenant()))
			if err != nil {
				evt.Nack(err)
				return events
			}
			thisTransform := deepcopy.Copy(f.config.Transformation)
			transform(subEvt, thisTransform, nil, "", -1)
			if isArray {
				evt.SetPathValue(fmt.Sprintf("%s[%d]", f.config.ToPath, idx), thisTransform, true)
			} else {
				evt.SetPathValue(f.config.ToPath, thisTransform, true)
			}
		}
		events = append(events, evt)
	}
	log.Ctx(evt.Context()).Debug().Str("op", "filter").Str("filterType", "transform").Str("name", f.Name()).Int("eventCount", len(events)).Msg("transform")
	return events
}

// transform is a helper function to perform a simple transformation
func transform(evt event.Event, t interface{}, parent interface{}, key string, idx int) {
	if evt == nil || t == nil {
		return
	}
	switch tt := t.(type) {
	case map[string]interface{}:
		for key, st := range tt {
			transform(evt, st, tt, key, -1)
		}
	case []interface{}:
		for idx, st := range tt {
			transform(evt, st, tt, "", idx)
		}
	case string:
		for {
			si := strings.Index(tt, "{")
			ei := strings.Index(tt, "}")
			if si < 0 || ei < 0 {
				break
			}
			path := tt[si+1 : ei]
			v, _, _ := evt.GetPathValue(path)
			v = deepcopy.Copy(v)
			if !(si == 0 && ei == len(tt)-1) {
				switch vt := v.(type) {
				case string:
					tt = tt[0:si] + vt + tt[ei+1:]
				default:
					sv, _ := json.Marshal(vt)
					tt = tt[0:si] + string(sv) + tt[ei+1:]
				}
				v = tt
			} else {
				tt = ""
			}
			if parent != nil {
				if key != "" {
					parent.(map[string]interface{})[key] = v
				} else if idx >= 0 {
					parent.([]interface{})[idx] = v
				}
			}
		}
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
