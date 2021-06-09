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

package block

import (
	"github.com/xmidt-org/ears/internal/pkg/rtsemconv"
	"github.com/xmidt-org/ears/pkg/event"
	"github.com/xmidt-org/ears/pkg/filter"
	"go.opentelemetry.io/otel"
)

var _ filter.Filterer = (*Filter)(nil)

func NewFilter(config interface{}) (*Filter, error) {
	return &Filter{}, nil
}

type Filter struct{}

// Filter lets no event pass
func (f *Filter) Filter(evt event.Event) []event.Event {
	if evt.Trace() {
		tracer := otel.Tracer(rtsemconv.EARSTracerName)
		_, span := tracer.Start(evt.Context(), "blockFilter")
		defer span.End()
	}
	evt.Ack()
	return []event.Event{}
}

func (f *Filter) Config() interface{} {
	if f == nil {
		return Config{}
	}
	return f.Config()
}

func (f *Filter) Name() string {
	return ""
}

func (f *Filter) Plugin() string {
	return "block"
}
