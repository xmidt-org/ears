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

package kinesis

import (
	"context"
	"encoding/json"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/credentials/stscreds"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/kinesis"
	"github.com/goccy/go-yaml"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"github.com/xmidt-org/ears/internal/pkg/rtsemconv"
	"github.com/xmidt-org/ears/internal/pkg/syncer"
	"github.com/xmidt-org/ears/pkg/event"
	"github.com/xmidt-org/ears/pkg/hasher"
	pkgplugin "github.com/xmidt-org/ears/pkg/plugin"
	"github.com/xmidt-org/ears/pkg/secret"
	"github.com/xmidt-org/ears/pkg/sender"
	"github.com/xmidt-org/ears/pkg/tenant"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/global"
	"go.opentelemetry.io/otel/metric/unit"
	"os"
	"time"
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
	ctx := context.Background()
	s := &Sender{
		name:            name,
		plugin:          plugin,
		tid:             tid,
		config:          cfg,
		logger:          event.GetEventLogger(),
		secrets:         secrets,
		awsRoleArn:      secrets.Secret(ctx, cfg.AWSRoleARN),
		awsAccessKey:    secrets.Secret(ctx, cfg.AWSAccessKeyId),
		awsAccessSecret: secrets.Secret(ctx, cfg.AWSSecretAccessKey),
		awsRegion:       secrets.Secret(ctx, cfg.AWSRegion),
		streamName:      secrets.Secret(ctx, cfg.StreamName),
		currentSec:      time.Now().Unix(),
		tableSyncer:     tableSyncer,
	}
	s.initPlugin()
	hostname, _ := os.Hostname()
	// metric recorders
	meter := global.Meter(rtsemconv.EARSMeterName)
	commonLabels := []attribute.KeyValue{
		attribute.String(rtsemconv.EARSPluginTypeLabel, rtsemconv.EARSPluginTypeKinesisSender),
		attribute.String(rtsemconv.EARSPluginNameLabel, s.Name()),
		attribute.String(rtsemconv.EARSAppIdLabel, s.tid.AppId),
		attribute.String(rtsemconv.EARSOrgIdLabel, s.tid.OrgId),
		attribute.String(rtsemconv.KinesisStreamNameLabel, s.streamName),
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

func (s *Sender) initPlugin() error {
	s.Lock()
	defer s.Unlock()
	sess, err := session.NewSession()
	if nil != err {
		return &KinesisError{op: "NewSession", err: err}
	}
	var creds *credentials.Credentials
	if s.awsRoleArn != "" {
		creds = stscreds.NewCredentials(sess, s.awsRoleArn)
	} else if s.awsAccessKey != "" && s.awsAccessSecret != "" {
		creds = credentials.NewStaticCredentials(s.awsAccessKey, s.awsAccessSecret, "")
	} else {
		creds = sess.Config.Credentials
	}
	sess, err = session.NewSession(&aws.Config{Region: aws.String(s.awsRegion), Credentials: creds})
	if nil != err {
		s.logError()
		return &KinesisError{op: "NewSession", err: err}
	}
	_, err = sess.Config.Credentials.Get()
	if nil != err {
		s.logError()
		return &KinesisError{op: "GetCredentials", err: err}
	}
	s.kinesisService = kinesis.New(sess)
	_, err = s.kinesisService.DescribeStream(&kinesis.DescribeStreamInput{StreamName: aws.String(s.streamName)})
	if err != nil {
		s.logError()
		return err
	}
	s.done = make(chan struct{})
	s.startTimedSender()
	return nil
}

func (s *Sender) startTimedSender() {
	go func() {
		for {
			select {
			case <-s.done:
				s.logger.Info().Str("op", "Kinesis.timedSender").Str("name", s.Name()).Str("tid", s.Tenant().ToString()).Str("app.id", s.Tenant().AppId).Str("partner.id", s.Tenant().OrgId).Msg("stopping kinesis sender")
				return
			case <-time.After(time.Duration(*s.config.SendTimeout) * time.Second):
			}
			s.Lock()
			if s.eventBatch == nil {
				s.eventBatch = make([]event.Event, 0)
			}
			evtBatch := s.eventBatch
			s.eventBatch = make([]event.Event, 0)
			s.Unlock()
			if len(evtBatch) > 0 {
				s.send(evtBatch)
			}
		}
	}()
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

func (s *Sender) send(events []event.Event) {
	if len(events) == 0 {
		return
	}
	batchReqs := []*kinesis.PutRecordsRequestEntry{}
	for idx, evt := range events {
		if idx == 0 {
			log.Ctx(evt.Context()).Debug().Str("op", "Kinesis.sendWorker").Str("name", s.Name()).Str("tid", s.Tenant().ToString()).Str("app.id", s.Tenant().AppId).Str("partner.id", s.Tenant().OrgId).Int("eventIdx", idx).Int("batchSize", len(events)).Msg("send message batch")
		}
		buf, err := json.Marshal(evt.Payload())
		if err != nil {
			s.logError()
			continue
		}
		partitionKey := s.config.PartitionKey
		if s.config.PartitionKeyPath != "" {
			pv, _, _ := evt.GetPathValue(s.config.PartitionKeyPath)
			pvstr, ok := pv.(string)
			if ok && pvstr != "" {
				partitionKey = pvstr
			}
		}
		if partitionKey == "" {
			partitionKey = uuid.New().String()
		}
		putReq := kinesis.PutRecordsRequestEntry{
			Data:         buf,
			PartitionKey: aws.String(partitionKey),
		}
		batchReqs = append(batchReqs, &putReq)
		s.eventBytesCounter.Add(evt.Context(), int64(len(buf)))
		s.eventProcessingTime.Record(evt.Context(), time.Since(evt.Created()).Milliseconds())
	}
	batchPut := kinesis.PutRecordsInput{
		Records:    batchReqs,
		StreamName: aws.String(s.streamName),
	}
	start := time.Now()
	putResults, err := s.kinesisService.PutRecordsWithContext(events[0].Context(), &batchPut)
	s.eventSendOutTime.Record(events[0].Context(), time.Since(start).Milliseconds())
	if err != nil {
		log.Ctx(events[0].Context()).Error().Str("op", "Kinesis.sendWorker").Str("name", s.Name()).Str("tid", s.Tenant().ToString()).Str("app.id", s.Tenant().AppId).Str("partner.id", s.Tenant().OrgId).Int("batchSize", len(events)).Msg("batch send error: " + err.Error())
		for idx := range events {
			s.logError()
			s.eventFailureCounter.Add(events[idx].Context(), 1)
			events[idx].Nack(err)
		}
	} else {
		for idx, putResult := range putResults.Records {
			if putResult.ErrorCode == nil {
				s.eventSuccessCounter.Add(events[idx].Context(), 1)
				s.logSuccess()
				events[idx].Ack()
			} else {
				s.eventFailureCounter.Add(events[idx].Context(), 1)
				s.logError()
				events[idx].Nack(err)
			}
		}
	}
}

func (s *Sender) Send(e event.Event) {
	s.Lock()
	if s.eventBatch == nil {
		s.eventBatch = make([]event.Event, 0)
	}
	s.eventBatch = append(s.eventBatch, e)
	if len(s.eventBatch) >= *s.config.MaxNumberOfMessages {
		eventBatch := s.eventBatch
		s.eventBatch = make([]event.Event, 0)
		s.Unlock()
		s.send(eventBatch)
	} else {
		s.Unlock()
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
