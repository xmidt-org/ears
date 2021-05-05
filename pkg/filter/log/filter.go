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

package log

import (
	"encoding/json"
	"github.com/rs/zerolog/log"
	"github.com/xmidt-org/ears/pkg/event"
	"github.com/xmidt-org/ears/pkg/filter"
)

var _ filter.Filterer = (*Filter)(nil)

func NewFilter(config interface{}) (*Filter, error) {
	return &Filter{}, nil
}

type Filter struct{}

// Filter log event and pass it on
func (f *Filter) Filter(evt event.Event) []event.Event {
	m := make(map[string]interface{}, 0)
	m["payload"] = evt.Payload()
	m["metadata"] = evt.Metadata()
	buf, err := json.Marshal(m)
	if err != nil {
		log.Ctx(evt.Context()).Error().Str("op", "log.Filter").Msg(err.Error())
	}
	log.Ctx(evt.Context()).Info().Str("op", "log.Filter").RawJSON("event", buf).Msg("ears log")
	return []event.Event{evt}
}
