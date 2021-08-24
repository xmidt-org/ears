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

package redis

import (
	"context"
	"encoding/json"
	"github.com/go-redis/redis"
	"github.com/goccy/go-yaml"
	"github.com/rs/zerolog/log"
	"github.com/xmidt-org/ears/internal/pkg/rtsemconv"
	"github.com/xmidt-org/ears/pkg/event"
	pkgplugin "github.com/xmidt-org/ears/pkg/plugin"
	"github.com/xmidt-org/ears/pkg/secret"
	"github.com/xmidt-org/ears/pkg/sender"
	"github.com/xmidt-org/ears/pkg/tenant"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/global"
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
		attribute.String(rtsemconv.EARSPluginTypeLabel, rtsemconv.EARSPluginTypeRedisSender),
		attribute.String(rtsemconv.EARSPluginNameLabel, s.Name()),
		attribute.String(rtsemconv.EARSAppIdLabel, s.tid.AppId),
		attribute.String(rtsemconv.EARSOrgIdLabel, s.tid.OrgId),
		attribute.String(rtsemconv.RedisChannelLabel, s.config.Channel),
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
	s.eventProcessingTime = metric.Must(meter).
		NewInt64ValueRecorder(
			rtsemconv.EARSMetricEventProcessingTime,
			metric.WithDescription("measures the number of event bytes processed"),
		).Bind(commonLabels...)
	return s, nil
}

func (s *Sender) initPlugin() error {
	s.Lock()
	s.client = redis.NewClient(&redis.Options{
		Addr:     s.config.Endpoint,
		Password: "",
		DB:       0,
	})
	s.Unlock()
	return nil
}

func (s *Sender) Count() int {
	s.Lock()
	defer s.Unlock()
	return s.count
}

func (s *Sender) StopSending(ctx context.Context) {
	s.Lock()
	s.client.Close()
	s.Unlock()
}

func (s *Sender) Send(e event.Event) {
	buf, err := json.Marshal(e.Payload())
	if err != nil {
		log.Ctx(e.Context()).Error().Str("op", "redis.Send").Msg("failed to marshal message: " + err.Error())
		s.eventFailureCounter.Add(e.Context(), 1)
		e.Nack(err)
		return
	}
	s.eventProcessingTime.Record(e.Context(), time.Since(e.Created()).Milliseconds())
	err = s.client.Publish(s.config.Channel, string(buf)).Err()
	if err != nil {
		log.Ctx(e.Context()).Error().Str("op", "redis.Send").Msg("failed to send message on redis channel: " + err.Error())
		s.eventFailureCounter.Add(e.Context(), 1)
		e.Nack(err)
		return
	}
	log.Ctx(e.Context()).Debug().Str("op", "redis.Send").Str("name", s.Name()).Str("tid", s.Tenant().ToString()).Msg("sent message on redis channel")
	s.eventSuccessCounter.Add(e.Context(), 1)
	s.Lock()
	s.count++
	s.Unlock()
	e.Ack()
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
