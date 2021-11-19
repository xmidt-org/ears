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

package kafka

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"github.com/Shopify/sarama"
	"github.com/goccy/go-yaml"
	"github.com/rs/zerolog/log"
	"github.com/xmidt-org/ears/internal/pkg/rtsemconv"
	"github.com/xmidt-org/ears/pkg/event"
	pkgplugin "github.com/xmidt-org/ears/pkg/plugin"
	"github.com/xmidt-org/ears/pkg/secret"
	"github.com/xmidt-org/ears/pkg/sender"
	"github.com/xmidt-org/ears/pkg/tenant"
	"go.opentelemetry.io/contrib/instrumentation/github.com/Shopify/sarama/otelsarama"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/global"
	"go.opentelemetry.io/otel/metric/unit"
	"strings"
	"time"
)

//TODO: support headers
//

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
		name:    name,
		plugin:  plugin,
		tid:     tid,
		config:  cfg,
		logger:  event.GetEventLogger(),
		secrets: secrets,
	}
	err = s.initPlugin()
	if err != nil {
		return nil, err
	}
	return s, nil
}

func (s *Sender) getLabelValues(e event.Event, labels []DynamicMetricLabel) []DynamicMetricValue {
	labelValues := make([]DynamicMetricValue, 0)
	for _, label := range labels {
		obj, _, _ := e.GetPathValue(label.Path)
		value, ok := obj.(string)
		if !ok {
			continue
		}
		labelValues = append(labelValues, DynamicMetricValue{Label: label.Label, Value: value})
	}
	return labelValues
}

func (s *Sender) getMetrics(labels []DynamicMetricValue) *SenderMetrics {
	s.Lock()
	defer s.Unlock()
	key := ""
	for _, label := range labels {
		key += label.Label + "-" + label.Value + "-"
	}
	if s.metrics == nil {
		s.metrics = make(map[string]*SenderMetrics)
	}
	m, ok := s.metrics[key]
	if !ok {
		newMetric := new(SenderMetrics)
		// metric recorders
		meter := global.Meter(rtsemconv.EARSMeterName)
		commonLabels := []attribute.KeyValue{
			attribute.String(rtsemconv.EARSPluginTypeLabel, rtsemconv.EARSPluginTypeKafkaSender),
			attribute.String(rtsemconv.EARSPluginNameLabel, s.Name()),
			attribute.String(rtsemconv.EARSAppIdLabel, s.tid.AppId),
			attribute.String(rtsemconv.EARSOrgIdLabel, s.tid.OrgId),
			attribute.String(rtsemconv.KafkaTopicLabel, s.config.Topic),
		}
		for _, label := range labels {
			commonLabels = append(commonLabels, attribute.String(label.Label, label.Value))
		}
		newMetric.eventSuccessCounter = metric.Must(meter).
			NewInt64Counter(
				rtsemconv.EARSMetricEventSuccess,
				metric.WithDescription("measures the number of successful events"),
			).Bind(commonLabels...)
		newMetric.eventFailureCounter = metric.Must(meter).
			NewInt64Counter(
				rtsemconv.EARSMetricEventFailure,
				metric.WithDescription("measures the number of unsuccessful events"),
			).Bind(commonLabels...)
		newMetric.eventBytesCounter = metric.Must(meter).
			NewInt64Counter(
				rtsemconv.EARSMetricEventBytes,
				metric.WithDescription("measures the number of event bytes processed"),
				metric.WithUnit(unit.Bytes),
			).Bind(commonLabels...)
		newMetric.eventProcessingTime = metric.Must(meter).
			NewInt64Histogram(
				rtsemconv.EARSMetricEventProcessingTime,
				metric.WithDescription("measures the time an event spends in ears"),
				metric.WithUnit(unit.Milliseconds),
			).Bind(commonLabels...)
		newMetric.eventSendOutTime = metric.Must(meter).
			NewInt64Histogram(
				rtsemconv.EARSMetricEventSendOutTime,
				metric.WithDescription("measures the time ears spends to send an event to a downstream data sink"),
				metric.WithUnit(unit.Milliseconds),
			).Bind(commonLabels...)
		s.metrics[key] = newMetric
		return newMetric
	} else {
		return m
	}
}

func (s *Sender) initPlugin() error {
	var err error
	s.Lock()
	s.producer, err = s.NewProducer(*s.config.SenderPoolSize)
	s.Unlock()
	if err != nil {
		return err
	}
	return nil
}

func (s *Sender) Count() int {
	s.Lock()
	defer s.Unlock()
	return s.count
}

func (s *Sender) StopSending(ctx context.Context) {
	s.Lock()
	defer s.Unlock()
	if !s.stopped {
		s.stopped = true
		for _, m := range s.metrics {
			m.eventSuccessCounter.Unbind()
			m.eventFailureCounter.Unbind()
			m.eventBytesCounter.Unbind()
			m.eventProcessingTime.Unbind()
			m.eventSendOutTime.Unbind()
		}
		s.producer.Close(ctx)
	}
}

func (s *Sender) NewProducer(count int) (*Producer, error) {
	if s.config.Brokers == "" {
		return nil, fmt.Errorf("Brokers cannot be empty")
	}
	ps, c, err := s.NewSyncProducers(count)
	if nil != err {
		return nil, err
	}
	pool := make(chan sarama.SyncProducer, count)
	for _, p := range ps {
		pool <- p
	}
	return &Producer{
		pool:   pool,
		done:   make(chan bool),
		client: c,
		logger: s.logger,
		sender: s,
	}, nil
}

func NewManualHashPartitioner(topic string) sarama.Partitioner {
	return &ManualHashPartitioner{
		Partitioner: sarama.NewHashPartitioner(topic),
	}
}

func (s *Sender) setConfig(config *sarama.Config) error {
	if "" != s.config.Version {
		v, err := sarama.ParseKafkaVersion(s.config.Version)
		if nil != err {
			return err
		}
		config.Version = v
	}
	if 0 < *s.config.ChannelBufferSize {
		config.ChannelBufferSize = *s.config.ChannelBufferSize
	}
	config.Net.TLS.Enable = s.config.TLSEnable
	if "" != s.config.Username {
		config.Net.TLS.Enable = true
		config.Net.SASL.Enable = true
		config.Net.SASL.User = s.config.Username
		config.Net.SASL.Password = s.secrets.Secret(s.config.Password)
		if config.Net.SASL.Password == "" {
			config.Net.SASL.Password = s.config.Password
		}
	} else if "" != s.config.AccessCert {
		accessCert := s.secrets.Secret(s.config.AccessCert)
		if accessCert == "" {
			accessCert = s.config.AccessCert
		}
		accessKey := s.secrets.Secret(s.config.AccessKey)
		if accessKey == "" {
			accessKey = s.config.AccessKey
		}
		caCert := s.secrets.Secret(s.config.CACert)
		if caCert == "" {
			caCert = s.config.CACert
		}
		keypair, err := tls.X509KeyPair([]byte(accessCert), []byte(accessKey))
		if err != nil {
			return err
		}
		caAuthorityPool := x509.NewCertPool()
		caAuthorityPool.AppendCertsFromPEM([]byte(caCert))
		tlsConfig := &tls.Config{
			Certificates: []tls.Certificate{keypair},
			RootCAs:      caAuthorityPool,
			MinVersion:   tls.VersionTLS12,
		}
		config.Net.TLS.Enable = true
		config.Net.TLS.Config = tlsConfig
	}
	return nil
}

func (s *Sender) NewSyncProducers(count int) ([]sarama.SyncProducer, sarama.Client, error) {
	brokers := strings.Split(s.config.Brokers, ",")
	config := sarama.NewConfig()
	config.Producer.Partitioner = NewManualHashPartitioner
	config.Producer.Compression = sarama.CompressionZSTD
	config.Producer.CompressionLevel = 1
	config.Producer.RequiredAcks = sarama.WaitForLocal //as long as one broker gets it
	config.Producer.Return.Successes = true
	if err := s.setConfig(config); nil != err {
		return nil, nil, err
	}
	producers := make([]sarama.SyncProducer, count)
	var client sarama.Client
	for i := 0; i < count; i++ {
		c, err := sarama.NewClient(brokers, config)
		if err != nil {
			return nil, nil, err
		}
		p, err := sarama.NewSyncProducerFromClient(c)
		if nil != err {
			c.Close()
			return nil, nil, err
		}

		//producers[i] = otelsarama.WrapSyncProducer(config, p)
		producers[i] = p
		client = c
	}
	return producers, client, nil
}

func (p *Producer) Partitions(topic string) ([]int32, error) {
	return p.client.Partitions(topic)
}

func (p *Producer) Close(ctx context.Context) {
	close(p.done)
	for i := 0; i < len(p.pool); i++ {
		producer := <-p.pool
		err := producer.Close()
		if err != nil {
			p.logger.Error().Str("op", "Producer.Close").Msg(err.Error())
		} else {
			p.logger.Info().Str("op", "Producer.Close").Msg("kafka producer closed")
		}
	}
}

func (p *Producer) SendMessage(ctx context.Context, topic string, partition int, headers map[string]string, bs []byte, e event.Event) error {
	hs := make([]sarama.RecordHeader, len(headers))
	idx := 0
	for k, v := range headers {
		hs[idx] = sarama.RecordHeader{
			Key:   []byte(k),
			Value: []byte(v),
		}
		idx++
	}
	var producer sarama.SyncProducer
	select {
	case <-p.done:
		//p.logger.Info().Str("op", "kafka.Send").Msg("producer done")
		return fmt.Errorf("producer closed")
	case producer = <-p.pool:
	}
	defer func() {
		p.pool <- producer
	}()
	start := time.Now()
	if partition != -1 {
		p.client.Partitions(topic)
		ps, err := p.client.Partitions(topic)
		if nil != err {
			return err
		}
		partition = partition % len(ps)
	}
	message := &sarama.ProducerMessage{
		Topic:     topic,
		Partition: int32(partition),
		Value:     sarama.StringEncoder(bs),
		Headers:   hs,
		Timestamp: time.Now(),
	}
	otel.GetTextMapPropagator().Inject(ctx, otelsarama.NewProducerMessageCarrier(message))

	part, offset, err := producer.SendMessage(message)
	if nil != err {
		return err
	}
	// override log values if any
	elapsed := time.Since(start).Milliseconds()
	p.sender.getMetrics(p.sender.getLabelValues(e, p.sender.config.DynamicMetricLabels)).eventSendOutTime.Record(ctx, elapsed)
	log.Ctx(ctx).Debug().Str("op", "kafka.Send").Int("elapsed", int(elapsed)).Int("partition", int(part)).Int("offset", int(offset)).Msg("sent message on kafka topic")
	return nil
}

func (mp *ManualHashPartitioner) Partition(message *sarama.ProducerMessage, numPartitions int32) (int32, error) {
	// manual partitioner
	if -1 < message.Partition {
		return message.Partition, nil
	}
	// hash partitioner
	return mp.Partitioner.Partition(message, numPartitions)
}

func (s *Sender) Send(e event.Event) {
	if s.stopped {
		log.Ctx(e.Context()).Error().Str("op", "kafka.Send").Str("name", s.Name()).Str("tid", s.Tenant().ToString()).Msg("drop message due to closed sender")
		s.getMetrics(s.getLabelValues(e, s.config.DynamicMetricLabels)).eventFailureCounter.Add(e.Context(), 1.0)
		return
	}
	buf, err := json.Marshal(e.Payload())
	if err != nil {
		log.Ctx(e.Context()).Error().Str("op", "kafka.Send").Str("name", s.Name()).Str("tid", s.Tenant().ToString()).Msg("failed to marshal message: " + err.Error())
		s.getMetrics(s.getLabelValues(e, s.config.DynamicMetricLabels)).eventFailureCounter.Add(e.Context(), 1.0)
		e.Nack(err)
		return
	}
	s.getMetrics(s.getLabelValues(e, s.config.DynamicMetricLabels)).eventProcessingTime.Record(e.Context(), time.Since(e.Created()).Milliseconds())
	s.getMetrics(s.getLabelValues(e, s.config.DynamicMetricLabels)).eventBytesCounter.Add(e.Context(), int64(len(buf)))
	partition := -1
	if s.config.PartitionPath != "" {
		val, _, _ := e.GetPathValue(s.config.PartitionPath)
		intVal, ok := val.(int)
		if ok {
			partition = intVal
		}
	} else {
		partition = *s.config.Partition
	}
	err = s.producer.SendMessage(e.Context(), s.config.Topic, partition, nil, buf, e)
	if err != nil {
		log.Ctx(e.Context()).Error().Str("op", "kafka.Send").Str("name", s.Name()).Str("tid", s.Tenant().ToString()).Msg("failed to send message: " + err.Error())
		s.getMetrics(s.getLabelValues(e, s.config.DynamicMetricLabels)).eventFailureCounter.Add(e.Context(), 1)
		e.Nack(err)
		return
	}
	s.getMetrics(s.getLabelValues(e, s.config.DynamicMetricLabels)).eventSuccessCounter.Add(e.Context(), 1)
	s.Lock()
	s.count++
	log.Ctx(e.Context()).Debug().Str("op", "kafka.Send").Str("name", s.Name()).Str("tid", s.Tenant().ToString()).Int("count", s.count).Msg("sent message on kafka topic")
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
