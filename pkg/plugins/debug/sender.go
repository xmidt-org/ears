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
	"encoding/json"
	"fmt"
	"github.com/xmidt-org/ears/internal/pkg/rtsemconv"
	"github.com/xmidt-org/ears/internal/pkg/syncer"
	"github.com/xmidt-org/ears/pkg/hasher"
	"github.com/xmidt-org/ears/pkg/secret"
	"github.com/xmidt-org/ears/pkg/tenant"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/global"
	"go.opentelemetry.io/otel/metric/unit"
	"time"

	"github.com/goccy/go-yaml"
	"github.com/xmidt-org/ears/pkg/event"
	pkgplugin "github.com/xmidt-org/ears/pkg/plugin"
	"github.com/xmidt-org/ears/pkg/sender"
)

func NewSender(tid tenant.Id, plugin string, name string, config interface{}, secrets secret.Vault, tableSyncer syncer.DeltaSyncer) (sender.Sender, error) {
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
		config:      cfg,
		name:        name,
		plugin:      plugin,
		tid:         tid,
		currentSec:  time.Now().Unix(),
		tableSyncer: tableSyncer,
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
			metric.WithUnit(unit.Bytes),
		).Bind(commonLabels...)
	s.eventProcessingTime = metric.Must(meter).
		NewInt64Histogram(
			rtsemconv.EARSMetricEventProcessingTime,
			metric.WithDescription("measures the time an event spends in ears"),
			metric.WithUnit(unit.Milliseconds),
		).Bind(commonLabels...)
	s.eventSendOutTime = metric.Must(meter).
		NewInt64Histogram(
			rtsemconv.EARSMetricEventSendOutTime,
			metric.WithDescription("measures the time ears spends to send an event to a downstream data sink"),
			metric.WithUnit(unit.Milliseconds),
		).Bind(commonLabels...)
	return s, nil
}

func (s *Sender) logSuccess() {
	s.Lock()
	s.successCounter++
	if time.Now().Unix() != s.currentSec {
		s.successVelocityCounter = s.currentSuccessVelocityCounter
		s.currentSuccessVelocityCounter = 0
		s.currentSec = time.Now().Unix()
	}
	s.currentSuccessVelocityCounter++
	s.Unlock()
}

func (s *Sender) logError() {
	s.Lock()
	s.errorCounter++
	if time.Now().Unix() != s.currentSec {
		s.errorVelocityCounter = s.currentErrorVelocityCounter
		s.currentErrorVelocityCounter = 0
		s.currentSec = time.Now().Unix()
	}
	s.currentErrorVelocityCounter++
	s.Unlock()
}

func (s *Sender) Send(e event.Event) {
	s.history.Add(e)
	buf, err := json.Marshal(e.Payload())
	if err != nil {
		s.eventFailureCounter.Add(e.Context(), 1)
		s.logError()
		e.Nack(err)
		return
	}
	s.eventBytesCounter.Add(e.Context(), int64(len(buf)))
	ept := time.Since(e.Created()).Milliseconds()
	s.eventProcessingTime.Record(e.Context(), ept)
	//fmt.Printf("SEND %p\n", e)
	if s.destination != nil {
		start := time.Now()
		err := s.destination.Write(e)
		est := time.Since(start).Milliseconds()
		s.eventSendOutTime.Record(e.Context(), est)
		if err != nil {
			s.eventFailureCounter.Add(e.Context(), 1)
			s.logError()
			e.Nack(err)
			return
		}
	}
	s.eventSuccessCounter.Add(e.Context(), 1)
	s.logSuccess()
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
	s.eventSuccessCounter.Unbind()
	s.eventFailureCounter.Unbind()
	s.eventBytesCounter.Unbind()
	s.eventProcessingTime.Unbind()
	s.eventSendOutTime.Unbind()
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

func (s *Sender) getLocalMetric() *syncer.EarsMetric {
	s.Lock()
	defer s.Unlock()
	metrics := &syncer.EarsMetric{
		s.successCounter,
		s.errorCounter,
		0,
		s.successVelocityCounter,
		s.errorVelocityCounter,
		0,
		s.currentSec,
	}
	return metrics
}

func (s *Sender) EventSuccessCount() int {
	hash := s.Hash()
	s.tableSyncer.WriteMetrics(hash, s.getLocalMetric())
	return s.tableSyncer.ReadMetrics(hash).SuccessCount
}

func (s *Sender) EventSuccessVelocity() int {
	hash := s.Hash()
	s.tableSyncer.WriteMetrics(hash, s.getLocalMetric())
	return s.tableSyncer.ReadMetrics(hash).SuccessVelocity
}

func (s *Sender) EventErrorCount() int {
	hash := s.Hash()
	s.tableSyncer.WriteMetrics(hash, s.getLocalMetric())
	return s.tableSyncer.ReadMetrics(hash).ErrorCount
}

func (s *Sender) EventErrorVelocity() int {
	hash := s.Hash()
	s.tableSyncer.WriteMetrics(hash, s.getLocalMetric())
	return s.tableSyncer.ReadMetrics(hash).ErrorVelocity
}

func (s *Sender) EventTs() int64 {
	hash := s.Hash()
	s.tableSyncer.WriteMetrics(hash, s.getLocalMetric())
	return s.tableSyncer.ReadMetrics(hash).LastEventTs
}

func (s *Sender) Hash() string {
	cfg := ""
	if s.Config() != nil {
		buf, _ := json.Marshal(s.Config())
		if buf != nil {
			cfg = string(buf)
		}
	}
	str := s.name + s.plugin + cfg
	hash := hasher.String(str)
	return hash
}
