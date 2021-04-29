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
	"errors"
	"fmt"
	"github.com/mohae/deepcopy"
	"github.com/xmidt-org/ears/pkg/event"
	"github.com/xmidt-org/ears/pkg/filter"
	"strings"
)

func NewFilter(config interface{}) (*Filter, error) {
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
	}
	return f, nil
}

// Filter splits an event containing an array into multiple events
func (f *Filter) Filter(evt event.Event) []event.Event {
	//TODO: add validation logic to filter
	//TODO: maybe replace with jq filter
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
		obj, _, _ := evt.GetPathValue(f.config.TransformPath, false)
		if obj == nil {
			evt.Nack(errors.New("no value at transform path " + f.config.TransformPath))
			return nil
		}
		obj = deepcopy.Copy(obj)
		transform(obj, thisTransform, nil, "", -1)
		evt.SetPathValue(f.config.ResultPath, thisTransform, false, true)
		events = append(events, evt)
	}
	return events
}

func (f *Filter) Config() Config {
	if f == nil {
		return Config{}
	}
	return f.config
}

// transform is a helper function to perform a simple transformation
func transform(a interface{}, t interface{}, parent interface{}, key string, idx int) {
	if a == nil || t == nil {
		return
	}
	switch t.(type) {
	case map[string]interface{}:
		for key, st := range t.(map[string]interface{}) {
			transform(a, st, t, key, -1)
		}
	case []interface{}:
		for idx, st := range t.([]interface{}) {
			transform(a, st, t, "", idx)
		}
	case string:
		expr := t.(string)
		if strings.HasPrefix(expr, "{") && strings.HasSuffix(expr, "}") {
			path := strings.Split(expr[1:len(expr)-1], ".")
			repl := a
			for _, p := range path {
				var ok bool
				repl, ok = repl.(map[string]interface{})[p]
				if !ok {
					break
				}
			}
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
