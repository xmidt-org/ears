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

package js

import (
	"github.com/rs/zerolog/log"
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
	return f, nil
}

// Filter executes javascript transformation
func (f *Filter) Filter(evt event.Event) []event.Event {
	transformedEvts, err := defaultInterpreter.Exec(evt, f.config.Source)
	if err != nil {
		log.Ctx(evt.Context()).Error().Str("op", "js.Filter").Msg("js filter error: " + err.Error())
		evt.Nack(err)
		return []event.Event{}
	}
	if transformedEvts == nil || len(transformedEvts) == 0 {
		evt.Ack()
		return []event.Event{}
	}
	return transformedEvts
}
