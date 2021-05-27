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

package batch

import (
	"fmt"
	"github.com/xmidt-org/ears/pkg/event"
	"github.com/xmidt-org/ears/pkg/filter"
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
	f.batch = make([]event.Event, 0)
	return f, nil
}

func (f *Filter) Filter(evt event.Event) []event.Event {
	if f == nil {
		evt.Nack(&filter.InvalidConfigError{
			Err: fmt.Errorf("<nil> pointer filter"),
		})
		return nil
	}
	f.Lock()
	defer f.Unlock()
	f.batch = append(f.batch, evt)
	if len(f.batch) >= *f.config.BatchSize {
		newEvt, err := evt.Clone(evt.Context())
		if err != nil {
			for _, e := range f.batch {
				e.Nack(err)
			}
			f.batch = make([]event.Event, 0)
			return []event.Event{}
		}
		batchPayload := make([]interface{}, 0)
		for _, e := range f.batch {
			batchPayload = append(batchPayload, e.Payload())
		}
		newEvt.SetPathValue("", batchPayload, true)
		for _, e := range f.batch {
			e.Ack()
		}
		f.batch = make([]event.Event, 0)
		return []event.Event{newEvt}
	}
	return []event.Event{}
}

func (f *Filter) Config() interface{} {
	if f == nil {
		return Config{}
	}
	return f.config
}

func (f *Filter) Name() string {
	return ""
}

func (f *Filter) Plugin() string {
	return "batch"
}
