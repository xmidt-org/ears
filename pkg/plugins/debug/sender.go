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

package debug

import (
	"context"
	"fmt"
	"github.com/xmidt-org/ears/internal/pkg/rtsemconv"
	"github.com/xmidt-org/ears/pkg/secret"
	"github.com/xmidt-org/ears/pkg/tenant"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/global"

	"github.com/goccy/go-yaml"
	"github.com/xmidt-org/ears/pkg/event"
	pkgplugin "github.com/xmidt-org/ears/pkg/plugin"
	"github.com/xmidt-org/ears/pkg/sender"
)

func NewSender(tid tenant.Id, plugin string, name string, config interface{}, secrets secret.Vault) (sender.Sender, error) {
	var cfg SenderConfig
	var err error
	switch c := config.(type) {
	case string:
		err = yaml.Unmarshal([]byte(c), &cfg)
	case []byte:
		err = yaml.Unmarshal(c, &cfg)
	case SenderConfig:
		cfg = c
	case *SenderConfig:
		cfg = *c
	}
	if err != nil {
		return nil, &pkgplugin.InvalidConfigError{
			Err: err,
		}
	}
	cfg = cfg.WithDefaults()
	err = cfg.Validate()
	if err != nil {
		return nil, err
	}
	s := &Sender{
		config: cfg,
		name:   name,
		plugin: plugin,
		tid:    tid,
	}
	switch s.config.Destination {
	case DestinationDevNull:
		s.destination = nil
	case DestinationStdout:
		s.destination = &SendStdout{}
	case DestinationStderr:
		s.destination = &SendStderr{}
	case DestinationCustom:
		s.destination = s.config.Writer
	default:
		return nil, &pkgplugin.InvalidConfigError{
			Err: fmt.Errorf("config.Destination value invalid"),
		}
	}
	s.history = newHistory(*s.config.MaxHistory)
	// metric recorders
	meter := global.Meter(rtsemconv.EARSMeterName)
	commonLabels := []attribute.KeyValue{
		attribute.String(rtsemconv.EARSPluginTypeLabel, rtsemconv.EARSPluginTypeDebugSender),
		attribute.String(rtsemconv.EARSPluginNameLabel, s.Name()),
		attribute.String(rtsemconv.EARSAppIdLabel, s.tid.AppId),
		attribute.String(rtsemconv.EARSOrgIdLabel, s.tid.OrgId),
	}
	s.eventSuccessCounter = metric.Must(meter).
		NewInt64Counter(
			rtsemconv.EARSMetricEventSuccess,
			metric.WithDescription("measures the number of successful events"),
		).Bind(commonLabels...)
	s.eventFailureCounter = metric.Must(meter).
		NewInt64Counter(
			rtsemconv.EARSMetricEventFailure,
			metric.WithDescription("measures the number of unsuccessful events"),
		).Bind(commonLabels...)
	s.eventBytesCounter = metric.Must(meter).
		NewInt64Counter(
			rtsemconv.EARSMetricEventBytes,
			metric.WithDescription("measures the number of event bytes processed"),
		).Bind(commonLabels...)
	return s, nil
}

func (s *Sender) Send(e event.Event) {
	s.history.Add(e)
	//fmt.Printf("SEND %p\n", e)
	if s.destination != nil {
		err := s.destination.Write(e)
		if err != nil {
			s.eventFailureCounter.Add(e.Context(), 1)
			e.Nack(err)
			return
		}
	}
	s.eventSuccessCounter.Add(e.Context(), 1)
	e.Ack()
}

func (s *Sender) Reset() {
	s.history.count = 0
}

func (s *Sender) Unwrap() sender.Sender {
	return s
}

func (s *Sender) Count() int {
	return s.history.Count()
}

func (s *Sender) History() []event.Event {
	history := s.history.History()
	events := make([]event.Event, len(history))
	for i, h := range history {
		if e, ok := h.(event.Event); ok {
			events[i] = e
		}
	}
	return events
}

func (s *Sender) StopSending(ctx context.Context) {
}

func (s *Sender) Config() interface{} {
	return s.config
}

func (s *Sender) Name() string {
	return s.name
}

func (s *Sender) Plugin() string {
	return s.plugin
}

func (s *Sender) Tenant() tenant.Id {
	return s.tid
}
