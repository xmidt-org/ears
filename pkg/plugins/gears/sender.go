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

package gears

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
	"github.com/xmidt-org/ears/internal/pkg/syncer"
	"github.com/xmidt-org/ears/pkg/event"
	"github.com/xmidt-org/ears/pkg/hasher"
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
	"go.opentelemetry.io/otel/trace"
	"hash/fnv"
	"math"
	"os"
	"sort"
	"strings"
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
	s := &Sender{
		name:    name,
		plugin:  plugin,
		tid:     tid,
		config:  cfg,
		logger:  event.GetEventLogger(),
		secrets: secrets,
	}
	s.MetricPlugin = pkgplugin.NewMetricPlugin(tableSyncer, s.Hash)
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
		hostname, _ := os.Hostname()
		newMetric := new(SenderMetrics)
		// metric recorders
		meter := global.Meter(rtsemconv.EARSMeterName)
		commonLabels := []attribute.KeyValue{
			attribute.String(rtsemconv.EARSPluginTypeLabel, rtsemconv.EARSPluginTypeGearsSender),
			attribute.String(rtsemconv.EARSPluginNameLabel, s.Name()),
			attribute.String(rtsemconv.EARSAppIdLabel, s.tid.AppId),
			attribute.String(rtsemconv.EARSOrgIdLabel, s.tid.OrgId),
			attribute.String(rtsemconv.KafkaTopicLabel, s.config.Topic),
			attribute.String(rtsemconv.HostnameLabel, hostname),
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
	s.Lock()
	defer s.Unlock()

	activeClusters := s.secrets.Secret(context.Background(), s.config.ActiveClusters)
	if len(activeClusters) > 0 {
		clusterKeys := strings.Split(activeClusters, ",")
		for _, key := range clusterKeys {
			settings, ok := s.config.Clusters[key]
			if !ok {
				return fmt.Errorf("activeClusters key %s not found in clusters", key)
			}
			producer, err := s.NewProducer(key, settings, *s.config.SenderPoolSize)
			if err != nil {
				return err
			}
			s.producers = append(s.producers, producer)
		}
	} else if len(s.config.Clusters) > 0 {
		for key, settings := range s.config.Clusters {
			producer, err := s.NewProducer(key, settings, *s.config.SenderPoolSize)
			if err != nil {
				return err
			}
			s.producers = append(s.producers, producer)
		}
	} else {
		producer, err := s.NewProducer("default", BrokerSettings{
			Brokers:    s.config.Brokers,
			Username:   s.config.Username,
			Password:   s.config.Password,
			CACert:     s.config.CACert,
			AccessCert: s.config.AccessCert,
			AccessKey:  s.config.AccessKey,
			TLSEnable:  s.config.TLSEnable,
		}, *s.config.SenderPoolSize)
		if err != nil {
			return err
		}
		s.producers = append(s.producers, producer)
	}
	sort.Slice(s.producers, func(i, j int) bool {
		return s.producers[i].key < s.producers[j].key
	})

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
		for _, p := range s.producers {
			p.Close(ctx)
		}
		s.DeleteMetrics()
	}
}

func (s *Sender) NewProducer(key string, settings BrokerSettings, count int) (*Producer, error) {
	if s.config.Brokers == "" {
		return nil, fmt.Errorf("Brokers cannot be empty")
	}
	ps, c, err := s.NewSyncProducers(settings, count)
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
		key:    key,
	}, nil
}

func NewManualHashPartitioner(topic string) sarama.Partitioner {
	return &ManualHashPartitioner{
		Partitioner: sarama.NewHashPartitioner(topic),
	}
}

func (s *Sender) setConfig(config *sarama.Config, settings BrokerSettings) error {
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
	config.Net.TLS.Enable = settings.TLSEnable
	ctx := context.Background()
	if "" != settings.Username {
		config.Net.TLS.Enable = true
		config.Net.SASL.Enable = true
		config.Net.SASL.User = settings.Username
		config.Net.SASL.Password = s.secrets.Secret(ctx, settings.Password)
		if config.Net.SASL.Password == "" {
			config.Net.SASL.Password = settings.Password
		}
	} else if "" != settings.AccessCert {
		accessCert := s.secrets.Secret(ctx, settings.AccessCert)
		if accessCert == "" {
			accessCert = s.config.AccessCert
		}
		accessKey := s.secrets.Secret(ctx, settings.AccessKey)
		if accessKey == "" {
			accessKey = settings.AccessKey
		}
		caCert := s.secrets.Secret(ctx, settings.CACert)
		if caCert == "" {
			caCert = settings.CACert
		}
		keypair, err := tls.X509KeyPair([]byte(accessCert), []byte(accessKey))
		if err != nil {
			return err
		}
		caAuthorityPool := x509.NewCertPool()
		caAuthorityPool.AppendCertsFromPEM([]byte(caCert))
		tlsConfig := &tls.Config{
			Certificates:       []tls.Certificate{keypair},
			RootCAs:            caAuthorityPool,
			MinVersion:         tls.VersionTLS12,
			InsecureSkipVerify: true,
		}
		config.Net.TLS.Enable = true
		config.Net.TLS.Config = tlsConfig
	}
	return nil
}

func (s *Sender) NewSyncProducers(settings BrokerSettings, count int) ([]sarama.SyncProducer, sarama.Client, error) {
	ctx := context.Background()
	brokersStr := s.secrets.Secret(ctx, settings.Brokers)
	if brokersStr == "" {
		brokersStr = settings.Brokers
	}
	brokers := strings.Split(brokersStr, ",")
	config := sarama.NewConfig()
	config.Producer.Partitioner = NewManualHashPartitioner
	config.Producer.RequiredAcks = sarama.WaitForLocal //as long as one broker gets it
	switch s.config.CompressionMethod {
	case "ZSTD":
		config.Producer.Compression = sarama.CompressionZSTD
	case "Snappy":
		config.Producer.Compression = sarama.CompressionSnappy
	case "GZIP":
		config.Producer.Compression = sarama.CompressionGZIP
	case "LZ4":
		config.Producer.Compression = sarama.CompressionLZ4
	}
	if s.config.CompressionLevel != nil {
		config.Producer.CompressionLevel = *s.config.CompressionLevel
	}
	config.Producer.Return.Successes = true
	if err := s.setConfig(config, settings); nil != err {
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
	log.Ctx(ctx).Debug().Str("op", "gears.Send").Str("topic", topic).Int("elapsed", int(elapsed)).Int("partition", int(part)).Int("offset", int(offset)).Msg("sent message on gears topic")
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
		log.Ctx(e.Context()).Error().Str("op", "gears.Send").Str("name", s.Name()).Str("tid", s.Tenant().ToString()).Msg("drop message due to closed sender")
		s.getMetrics(s.getLabelValues(e, s.config.DynamicMetricLabels)).eventFailureCounter.Add(e.Context(), 1.0)
		return
	}
	if s.config.Enveloped {
		location := ""
		obj, _, _ := e.Evaluate("{.to.location}")
		if obj != nil {
			switch l := obj.(type) {
			case string:
				location = l
			}
		}
		if location == "" {
			log.Ctx(e.Context()).Error().Str("op", "gears.Send").Str("name", s.Name()).Str("tid", s.Tenant().ToString()).Msg("invalid gears envelope missing location")
			if span := trace.SpanFromContext(e.Context()); span != nil {
				span.AddEvent("invalid gears envelope")
			}
			e.Ack()
			return
		}
		//if tx.traceId not set and if e.UserTraceId() is set, set tx.traceId to e.UserTraceId()
		payload, ok := e.Payload().(map[string]interface{})
		if ok {
			tx, ok := payload["tx"].(map[string]interface{})
			if ok {
				_, ok := tx["traceId"]
				if !ok && e.UserTraceId() != "" {
					tx["traceId"] = e.UserTraceId()
					payload["tx"] = tx
				}
			} else {
				if e.UserTraceId() != "" {
					payload["tx"] = map[string]interface{}{"traceId": e.UserTraceId()}
				}
			}
		}

		buf, err := json.Marshal(payload)
		if err != nil {
			log.Ctx(e.Context()).Error().Str("op", "gears.Send").Str("name", s.Name()).Str("tid", s.Tenant().ToString()).Msg("failed to marshal message: " + err.Error())
			s.getMetrics(s.getLabelValues(e, s.config.DynamicMetricLabels)).eventFailureCounter.Add(e.Context(), 1.0)
			s.LogError()
			e.Nack(err)
			return
		}
		s.getMetrics(s.getLabelValues(e, s.config.DynamicMetricLabels)).eventProcessingTime.Record(e.Context(), time.Since(e.Created()).Milliseconds())
		s.getMetrics(s.getLabelValues(e, s.config.DynamicMetricLabels)).eventBytesCounter.Add(e.Context(), int64(len(buf)))
		hashbuf := []byte(location)
		h := fnv.New32a()
		h.Write(hashbuf)
		pIdx := getProducerIdx(h.Sum32(), len(s.producers))
		partition := int(h.Sum32())
		err = s.producers[pIdx].SendMessage(e.Context(), s.config.Topic, partition, nil, buf, e)
		if err != nil {
			log.Ctx(e.Context()).Error().Str("op", "gears.Send").Str("name", s.Name()).Str("tid", s.Tenant().ToString()).Msg("failed to send message: " + err.Error())
			s.getMetrics(s.getLabelValues(e, s.config.DynamicMetricLabels)).eventFailureCounter.Add(e.Context(), 1)
			s.LogError()
			e.Nack(err)
			return
		}
		s.getMetrics(s.getLabelValues(e, s.config.DynamicMetricLabels)).eventSuccessCounter.Add(e.Context(), 1)
		s.LogSuccess()
		s.Lock()
		s.count++
		log.Ctx(e.Context()).Info().Str("op", "gears.Send").
			Str("name", s.Name()).Uint32("hash", h.Sum32()).Str("producer", s.producers[pIdx].key).
			Str("location", location).Str("tid", s.Tenant().ToString()).Int("count", s.count).
			Msg("sent enveloped message on gears topic")
		s.Unlock()
	} else {
		to := make(map[string]string, 0)
		obj, _, _ := e.Evaluate(s.config.Partner)
		if obj != nil {
			switch partner := obj.(type) {
			case string:
				to["partner"] = partner
			}
		}
		if to["partner"] == "" {
			log.Ctx(e.Context()).Error().Str("op", "gears.Send").Str("name", s.Name()).Str("tid", s.Tenant().ToString()).Msg("missing gears routing information partner")
			if span := trace.SpanFromContext(e.Context()); span != nil {
				span.AddEvent("missing partner")
			}
			e.Ack()
			return
		}
		obj, _, _ = e.Evaluate(s.config.App)
		if obj != nil {
			switch app := obj.(type) {
			case string:
				to["app"] = app
			}
		}
		if to["app"] == "" {
			log.Ctx(e.Context()).Error().Str("op", "gears.Send").Str("name", s.Name()).Str("tid", s.Tenant().ToString()).Msg("missing gears routing information app")
			if span := trace.SpanFromContext(e.Context()); span != nil {
				span.AddEvent("missing app")
			}
			e.Ack()
			return
		}
		locations := make([]string, 0)
		obj, _, _ = e.Evaluate(s.config.Location)
		if obj != nil {
			switch loc := obj.(type) {
			case string:
				locations = append(locations, loc)
			case []string:
				for _, ls := range loc {
					obj, _, _ := e.Evaluate(ls)
					switch sls := obj.(type) {
					case string:
						locations = append(locations, sls)
					}
				}
			case []interface{}:
				for _, l := range loc {
					if ls, ok := l.(string); ok {
						obj, _, _ := e.Evaluate(ls)
						switch sls := obj.(type) {
						case string:
							locations = append(locations, sls)
						}
					}
				}
			}
		}
		if len(locations) == 0 {
			log.Ctx(e.Context()).Error().Str("op", "gears.Send").Str("name", s.Name()).Str("tid", s.Tenant().ToString()).Msg("missing gears routing information location")
			if span := trace.SpanFromContext(e.Context()); span != nil {
				span.AddEvent("missing location")
			}
			e.Ack()
			return
		}
		tx := make(map[string]string, 0)
		if e.UserTraceId() != "" {
			tx["traceId"] = e.UserTraceId()
		} else {
			obj, _, _ = e.Evaluate("{trace.id}")
			if obj != nil {
				switch txid := obj.(type) {
				case string:
					tx["traceId"] = txid
				}
			}
		}
		message := make(map[string]interface{}, 0)
		message["op"] = "process"
		if s.config.Uses != "" {
			message["uses"] = s.config.Uses
		}
		message["payload"] = e.Payload()
		envelope := make(map[string]interface{}, 0)
		envelope["message"] = message
		envelope["to"] = to
		envelope["tx"] = tx
		for _, l := range locations {
			to["location"] = l
			buf, err := json.Marshal(envelope)
			if err != nil {
				log.Ctx(e.Context()).Error().Str("op", "gears.Send").Str("name", s.Name()).Str("tid", s.Tenant().ToString()).Msg("failed to marshal message: " + err.Error())
				s.getMetrics(s.getLabelValues(e, s.config.DynamicMetricLabels)).eventFailureCounter.Add(e.Context(), 1.0)
				s.LogError()
				e.Nack(err)
				return
			}
			s.getMetrics(s.getLabelValues(e, s.config.DynamicMetricLabels)).eventProcessingTime.Record(e.Context(), time.Since(e.Created()).Milliseconds())
			s.getMetrics(s.getLabelValues(e, s.config.DynamicMetricLabels)).eventBytesCounter.Add(e.Context(), int64(len(buf)))
			hashbuf := []byte(to["location"])
			h := fnv.New32a()
			h.Write(hashbuf)
			pIdx := getProducerIdx(h.Sum32(), len(s.producers))
			partition := int(h.Sum32())
			err = s.producers[pIdx].SendMessage(e.Context(), s.config.Topic, partition, nil, buf, e)
			if err != nil {
				log.Ctx(e.Context()).Error().Str("op", "gears.Send").Str("name", s.Name()).Str("tid", s.Tenant().ToString()).Msg("failed to send message: " + err.Error())
				s.getMetrics(s.getLabelValues(e, s.config.DynamicMetricLabels)).eventFailureCounter.Add(e.Context(), 1)
				s.LogError()
				e.Nack(err)
				return
			}
			s.getMetrics(s.getLabelValues(e, s.config.DynamicMetricLabels)).eventSuccessCounter.Add(e.Context(), 1)
			s.LogSuccess()
			s.Lock()
			s.count++
			log.Ctx(e.Context()).Info().Str("op", "gears.Send").Str("name", s.Name()).Uint32("hash", h.Sum32()).Str("producer", s.producers[pIdx].key).Str("location", l).Str("tid", s.Tenant().ToString()).Int("count", s.count).Msg("sent message on gears topic")
			s.Unlock()
		}
	}
	e.Ack()
}

// Given a hash and the number of producers, return the index of the producer to use
// We do not use % here because we are going to % the hash again in SendMessage
// Instead, we index by dividing uint32 by the number of producers and see which
// bucket is the hash in.
func getProducerIdx(hash uint32, count int) int {
	bucket := math.MaxUint32 / uint32(count)
	return int(hash / bucket)
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
