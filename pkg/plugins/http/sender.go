// Copyright 2021 Comcast Cable Communications Management, LLC
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

package http

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/goccy/go-yaml"
	"github.com/xmidt-org/ears/internal/pkg/rtsemconv"
	"github.com/xmidt-org/ears/internal/pkg/syncer"
	"github.com/xmidt-org/ears/pkg/event"
	"github.com/xmidt-org/ears/pkg/hasher"
	pkgplugin "github.com/xmidt-org/ears/pkg/plugin"
	"github.com/xmidt-org/ears/pkg/secret"
	"github.com/xmidt-org/ears/pkg/sender"
	"github.com/xmidt-org/ears/pkg/tenant"
	"go.opentelemetry.io/contrib/propagators/b3"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/global"
	"go.opentelemetry.io/otel/metric/unit"
	"go.opentelemetry.io/otel/propagation"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"time"
)

const DEFAULT_TIMEOUT = 10

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
	err = cfg.Validate()
	if err != nil {
		return nil, err
	}
	//TODO Does this live here?
	//TODO Make this a configuration?
	http.DefaultTransport.(*http.Transport).MaxIdleConnsPerHost = 100
	s := &Sender{
		client: &http.Client{
			Timeout: DEFAULT_TIMEOUT * time.Second,
		},
		config: cfg,
		name:   name,
		plugin: plugin,
		tid:    tid,
	}
	s.MetricPlugin = pkgplugin.NewMetricPlugin(tableSyncer)
	// metric recorders
	hostname, _ := os.Hostname()
	meter := global.Meter(rtsemconv.EARSMeterName)
	commonLabels := []attribute.KeyValue{
		attribute.String(rtsemconv.EARSPluginTypeLabel, rtsemconv.EARSPluginTypeHttpSender),
		attribute.String(rtsemconv.EARSPluginNameLabel, s.Name()),
		attribute.String(rtsemconv.EARSAppIdLabel, s.tid.AppId),
		attribute.String(rtsemconv.EARSOrgIdLabel, s.tid.OrgId),
		attribute.String(rtsemconv.HostnameLabel, hostname),
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

	s.b3Propagator = b3.New()
	return s, nil
}

func (s *Sender) Send(event event.Event) {
	payload := event.Payload()
	body, err := json.Marshal(payload)
	if err != nil {
		s.eventFailureCounter.Add(event.Context(), 1)
		s.LogError()
		event.Nack(err)
		return
	}
	s.eventBytesCounter.Add(event.Context(), int64(len(body)))
	s.eventProcessingTime.Record(event.Context(), time.Since(event.Created()).Milliseconds())
	req, err := http.NewRequest(s.config.Method, s.config.Url, bytes.NewReader(body))
	if err != nil {
		s.eventFailureCounter.Add(event.Context(), 1)
		s.LogError()
		event.Nack(err)
		return
	}
	ctx := event.Context()
	s.b3Propagator.Inject(ctx, propagation.HeaderCarrier(req.Header))
	start := time.Now()
	resp, err := s.client.Do(req)
	s.eventSendOutTime.Record(event.Context(), time.Since(start).Milliseconds())
	if err != nil {
		s.eventFailureCounter.Add(event.Context(), 1)
		s.LogError()
		event.Nack(err)
		return
	}
	io.Copy(ioutil.Discard, resp.Body)
	defer resp.Body.Close()
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusBadRequest {
		s.eventFailureCounter.Add(event.Context(), 1)
		s.LogError()
		event.Nack(&BadHttpStatusError{resp.StatusCode})
		return
	}
	s.eventSuccessCounter.Add(event.Context(), 1)
	s.LogSuccess()
	event.Ack()
}

func (s *Sender) Unwrap() sender.Sender {
	return nil
}

func (s *Sender) StopSending(ctx context.Context) {
	s.eventSuccessCounter.Unbind()
	s.eventFailureCounter.Unbind()
	s.eventBytesCounter.Unbind()
	s.eventProcessingTime.Unbind()
	s.eventSendOutTime.Unbind()
}

func (r *Sender) Config() interface{} {
	return r.config
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
