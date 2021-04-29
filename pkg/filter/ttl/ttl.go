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
	"github.com/xmidt-org/ears/pkg/event"
	"github.com/xmidt-org/ears/pkg/filter"
	"time"
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
	if f == nil {
		evt.Nack(&filter.InvalidConfigError{
			Err: fmt.Errorf("<nil> pointer filter"),
		})
		return nil
	}
	obj, _, _ := evt.GetPathValue(f.config.TtlPath, false)
	if obj == nil {
		evt.Nack(errors.New("nil object at unwrap path " + f.config.TtlPath))
		return []event.Event{}
	}
	evtTs, ok := obj.(uint64)
	if !ok {
		evt.Nack(errors.New("not a uint64 at ttl path " + f.config.TtlPath))
		return []event.Event{}
	}
	nowNanos := time.Now().UnixNano()
	if int64(evtTs) > nowNanos {
		evt.Nack(errors.New("event from the future at ttl path " + f.config.TtlPath))
		return []event.Event{}
	}
	if nowNanos-int64(evtTs) >= int64(*f.config.Ttl*(*f.config.TtlNanoFactor)) {
		evt.Ack()
		return []event.Event{}
	}
	return []event.Event{evt}
}
func (f *Filter) Config() Config {
	if f == nil {
		return Config{}
	}
	return f.config
}
