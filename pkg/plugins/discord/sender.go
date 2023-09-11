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

package discord

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/xmidt-org/ears/internal/pkg/syncer"
	"github.com/xmidt-org/ears/pkg/hasher"
	"os"
	"time"

	"github.com/bwmarrin/discordgo"
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
	s.initPlugin()
	hostname, _ := os.Hostname()
	// metric recorders
	meter := global.Meter(rtsemconv.EARSMeterName)
	commonLabels := []attribute.KeyValue{
		attribute.String(rtsemconv.EARSPluginTypeLabel, rtsemconv.EARSPluginTypeDiscordSender),
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

func (s *Sender) Send(event event.Event) {
	payload := event.Payload()
	content, ok := payload.(map[string]interface{})["content"].(string)
	if !ok {
		s.eventFailureCounter.Add(event.Context(), 1)
		s.logError()
		event.Nack(errors.New("Bad input for discord message"))
		return
	}
	var embeds []*discordgo.MessageEmbed
	message := &discordgo.MessageSend{Content: content, Embeds: embeds}
	s.eventBytesCounter.Add(event.Context(), int64(len(content)))
	s.eventProcessingTime.Record(event.Context(), time.Since(event.Created()).Milliseconds())
	_, err := s.sess.ChannelMessageSendComplex(s.config.ChannelId, message)
	if err != nil {
		s.eventFailureCounter.Add(event.Context(), 1)
		s.logError()
		event.Nack(err)
		return
	}
	s.eventSuccessCounter.Add(event.Context(), 1)
	s.logSuccess()
	event.Ack()
}

func (s *Sender) initPlugin() error {
	sess, err := discordgo.New("Bot " + s.config.BotToken)
	if nil != err {
		return err
	}
	s.sess = sess
	s.sess.Identify.Shard = &[2]int{0, 1}
	return sess.Open()
}

func (s *Sender) StopSending(ctx context.Context) {
	s.eventSuccessCounter.Unbind()
	s.eventFailureCounter.Unbind()
	s.eventBytesCounter.Unbind()
	s.eventProcessingTime.Unbind()
	s.eventSendOutTime.Unbind()
}

func (s *Sender) Unwrap() sender.Sender {
	return s
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
		0,
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
