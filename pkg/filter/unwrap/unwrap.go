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

package unwrap

import (
	"context"
	"errors"
	"fmt"
	"github.com/xmidt-org/ears/pkg/event"
	"github.com/xmidt-org/ears/pkg/filter"
	"strings"
)

// a SplitFilter splits and event into two or more events
/*type SplitFilter struct {
	SplitPath string // path to split array in payload
}*/

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
func (f *Filter) Filter(ctx context.Context, evt event.Event) ([]event.Event, error) {
	//TODO: add validation logic to filter
	//TODO: maybe replace with jq filter
	if f == nil {
		return nil, &filter.InvalidConfigError{
			Err: fmt.Errorf("<nil> pointer filter"),
		}
	}
	events := []event.Event{}
	if f.config.UnwrapPath == "" {
		events = append(events, evt)
	} else if f.config.UnwrapPath == "." {
		events = append(events, evt)
	} else {
		path := strings.Split(f.config.UnwrapPath, ".")
		obj := evt.Payload()
		for _, p := range path {
			var ok bool
			obj, ok = obj.(map[string]interface{})[p]
			if !ok {
				return events, errors.New("invalid object in filter path")
			}
		}
		nevt, err := evt.Dup()
		if err != nil {
			return events, err
		}
		err = nevt.SetPayload(obj)
		if err != nil {
			return events, err
		}
		events = append(events, nevt)
	}
	return events, nil
}

func (f *Filter) Config() Config {
	if f == nil {
		return Config{}
	}
	return f.config
}
