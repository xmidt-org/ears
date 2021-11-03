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

package s3

import (
	"context"
	"encoding/json"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/goccy/go-yaml"
	"github.com/xmidt-org/ears/internal/pkg/rtsemconv"
	"github.com/xmidt-org/ears/pkg/event"
	pkgplugin "github.com/xmidt-org/ears/pkg/plugin"
	"github.com/xmidt-org/ears/pkg/secret"
	"github.com/xmidt-org/ears/pkg/sender"
	"github.com/xmidt-org/ears/pkg/tenant"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/global"
	"go.opentelemetry.io/otel/metric/unit"
	"strings"
	"time"
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
		name:   name,
		plugin: plugin,
		tid:    tid,
		config: cfg,
		logger: event.GetEventLogger(),
	}
	s.initPlugin()
	// metric recorders
	meter := global.Meter(rtsemconv.EARSMeterName)
	commonLabels := []attribute.KeyValue{
		attribute.String(rtsemconv.EARSPluginTypeLabel, rtsemconv.EARSPluginTypeSQSSender),
		attribute.String(rtsemconv.EARSPluginNameLabel, s.Name()),
		attribute.String(rtsemconv.EARSAppIdLabel, s.tid.AppId),
		attribute.String(rtsemconv.EARSOrgIdLabel, s.tid.OrgId),
		attribute.String(rtsemconv.S3Bucket, s.config.Bucket),
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

func (s *Sender) initPlugin() error {
	s.Lock()
	defer s.Unlock()
	var err error
	s.session, err = session.NewSession(&aws.Config{
		Region: aws.String(s.config.Region),
	})
	if nil != err {
		return err
	}
	_, err = s.session.Config.Credentials.Get()
	if nil != err {
		return err
	}
	s.s3Service = s3.New(s.session)
	s.done = make(chan struct{})
	return nil
}

func (s *Sender) Count() int {
	s.Lock()
	defer s.Unlock()
	return s.count
}

func (s *Sender) StopSending(ctx context.Context) {
	s.Lock()
	if s.done != nil {
		s.eventSuccessCounter.Unbind()
		s.eventFailureCounter.Unbind()
		s.eventBytesCounter.Unbind()
		s.eventProcessingTime.Unbind()
		s.eventSendOutTime.Unbind()
		s.done <- struct{}{}
		s.done = nil
	}
	s.Unlock()
}

func (s *Sender) Send(evt event.Event) {
	buf, err := json.Marshal(evt.Payload())
	if err != nil {
		s.eventFailureCounter.Add(evt.Context(), 1)
		evt.Nack(err)
		return
	}
	s.eventBytesCounter.Add(evt.Context(), int64(len(buf)))
	s.eventProcessingTime.Record(evt.Context(), time.Since(evt.Created()).Milliseconds())
	start := time.Now()
	uploader := s3manager.NewUploader(s.session)
	_, err = uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(s.config.Bucket),
		Key:    aws.String(s.config.Path),
		Body:   strings.NewReader(string(buf)),
	})
	s.eventSendOutTime.Record(evt.Context(), time.Since(start).Milliseconds())
	if err != nil {
		s.eventFailureCounter.Add(evt.Context(), 1)
		evt.Nack(err)
	} else {
		s.Lock()
		s.count++
		s.Unlock()
		s.eventSuccessCounter.Add(evt.Context(), 1)
		evt.Ack()
	}
}

func (s *Sender) Unwrap() sender.Sender {
	return s
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
