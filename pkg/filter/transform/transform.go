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
func (f *Filter) Filter(evt event.Event) ([]event.Event, error) {
	//TODO: add validation logic to filter
	//TODO: maybe replace with jq filter
	if f == nil {
		return nil, &filter.InvalidConfigError{
			Err: fmt.Errorf("<nil> pointer filter"),
		}
	}
	events := []event.Event{}
	if f.config.Transformation == nil {
		events = append(events, evt)
	} else {
		thisTransfrom := deepcopy.Copy(f.config.Transformation)
		transform(evt.Payload(), thisTransfrom, nil, "", -1)
		err := evt.SetPayload(thisTransfrom)
		if err != nil {
			return events, err
		}
		events = append(events, evt)
	}
	return events, nil
}

func (f *Filter) Config() Config {
	if f == nil {
		return Config{}
	}
	return f.config
}

// transform is a helper function to perfrom a simple transformation
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
