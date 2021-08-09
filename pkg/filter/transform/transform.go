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
		thisTransform := deepcopy.Copy(f.config.Transformation)
		transform(evt, thisTransform, nil, "", -1)
		evt.SetPathValue(f.config.ToPath, thisTransform, true)
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
	switch t := t.(type) {
	case map[string]interface{}:
		for key, st := range t {
			transform(evt, st, t, key, -1)
		}
	case []interface{}:
		for idx, st := range t {
			transform(evt, st, t, "", idx)
		}
	case string:
		if strings.HasPrefix(t, "{") && strings.HasSuffix(t, "}") {
			repl, _, _ := evt.GetPathValue(t[1 : len(t)-1])
			if parent != nil {
				if key != "" {
					parent.(map[string]interface{})[key] = repl
				} else if idx >= 0 {
					parent.([]interface{})[idx] = repl
				}
			}
		}
	}
}

func (f *Filter) Config() interface{} {
	if f == nil {
		return Config{}
	}
	return f.Config()
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
