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

package decode

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
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

func (f *Filter) Filter(evt event.Event) []event.Event {
	//TODO: add validation logic to filter
	if f == nil {
		evt.Nack(&filter.InvalidConfigError{
			Err: fmt.Errorf("<nil> pointer filter"),
		})
		return nil
	}
	obj := evt.Payload()
	var parent interface{}
	var key string
	if f.config.DecodePath == "" {
	} else if f.config.DecodePath == "." {
	} else {
		path := strings.Split(f.config.DecodePath, ".")
		for _, p := range path {
			var ok bool
			parent = obj
			key = p
			obj, ok = obj.(map[string]interface{})[p]
			if !ok {
				evt.Nack(errors.New("invalid object at decode path " + f.config.DecodePath))
				return []event.Event{}
			}
		}
	}
	if obj == nil {
		evt.Nack(errors.New("nil object at decode path " + f.config.DecodePath))
		return []event.Event{}
	}
	var input string
	switch obj.(type) {
	case string:
		input = obj.(string)
	case []byte:
		input = string(obj.([]byte))
	default:
		evt.Nack(errors.New("unsupported field type at decode path " + f.config.DecodePath))
		return []event.Event{}
	}
	buf, err := base64.StdEncoding.DecodeString(input)
	if err != nil {
		evt.Nack(err)
		return []event.Event{}
	}
	var output interface{}
	err = json.Unmarshal(buf, &output)
	if err != nil {
		evt.Nack(err)
		return []event.Event{}
	}
	parentMap, is := parent.(map[string]interface{})
	if !is {
		evt.Nack(errors.New("parent is not a map at decode path " + f.config.DecodePath))
		return []event.Event{}
	}
	if key != "" && parentMap != nil {
		parentMap[key] = output
	} else {
		evt.SetPayload(output)
	}
	return []event.Event{evt}
}
func (f *Filter) Config() Config {
	if f == nil {
		return Config{}
	}
	return f.config
}
