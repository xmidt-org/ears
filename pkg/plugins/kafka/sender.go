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
	"github.com/rs/zerolog"
	"github.com/xmidt-org/ears/pkg/event"
	pkgplugin "github.com/xmidt-org/ears/pkg/plugin"
	"github.com/xmidt-org/ears/pkg/secret"
	"github.com/xmidt-org/ears/pkg/sender"
	"github.com/xmidt-org/ears/pkg/tenant"
	"os"
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
	logger := zerolog.New(os.Stdout).Level(zerolog.DebugLevel).With().Timestamp().Logger()
	//zerolog.LevelFieldName = "log.level"
	s := &Sender{
		name:    name,
		plugin:  plugin,
		tid:     tid,
		config:  cfg,
		logger:  logger,
		secrets: secrets,
	}

	err = s.initPlugin()
	if err != nil {
		return nil, err
	}
	return s, nil
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
	config.Producer.RequiredAcks = sarama.WaitForLocal //as long as one broker gets it
	config.Producer.Return.Successes = true
	if err := s.setConfig(config); nil != err {
		return nil, nil, err
	}
	ret := make([]sarama.SyncProducer, count)
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
		ret[i] = p
		client = c
	}
	return ret, client, nil
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

func (p *Producer) SendMessage(topic string, partition int, headers map[string]string, bs []byte) error {
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
	part, offset, err := producer.SendMessage(message)
	if nil != err {
		return err
	}
	// override log values if any
	elapsed := time.Since(start)
	p.logger.Info().Str("op", "kafka.Send").Int("elapsed", int(elapsed.Milliseconds())).Int("partition", int(part)).Int("offset", int(offset)).Msg("sent message on kafka topic")
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
		s.logger.Info().Str("op", "kafka.Send").Msg("drop message due to closed sender")
		return
	}
	buf, err := json.Marshal(e.Payload())
	if err != nil {
		s.logger.Error().Str("op", "kafka.Send").Msg("failed to marshal message: " + err.Error())
		e.Nack(err)
		return
	}
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
	err = s.producer.SendMessage(s.config.Topic, partition, nil, buf)
	if err != nil {
		s.logger.Error().Str("op", "kafka.Send").Msg("failed to send message: " + err.Error())
		e.Nack(err)
		return
	}
	s.Lock()
	s.count++
	s.logger.Info().Str("op", "kafka.Send").Int("count", s.count).Msg("sent message on kafka topic")
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
