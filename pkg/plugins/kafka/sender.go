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
	"github.com/xmidt-org/ears/pkg/sender"
	"os"
	"strings"
	"time"
)

func NewSender(config interface{}) (sender.Sender, error) {
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
	logger := zerolog.New(os.Stdout).Level(zerolog.DebugLevel)
	zerolog.LevelFieldName = "log.level"
	s := &Sender{
		config: cfg,
		logger: logger,
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
	s.Unlock()
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
		config.Net.SASL.Password = s.config.Password
	} else if "" != s.config.AccessCert {
		keypair, err := tls.X509KeyPair([]byte(s.config.AccessCert), []byte(s.config.AccessKey))
		if err != nil {
			return err
		}
		caAuthorityPool := x509.NewCertPool()
		caAuthorityPool.AppendCertsFromPEM([]byte(s.config.CACert))
		tlsConfig := &tls.Config{
			Certificates: []tls.Certificate{keypair},
			RootCAs:      caAuthorityPool,
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

func (p *Producer) Close(ctx *context.Context) {
	close(p.done)
	for i := 0; i < len(p.pool); i++ {
		producer := <-p.pool
		if err := producer.Close(); nil != err {
			//ctx.Log.Error("op", "Producer.Close", "error", err)
		}
	}
}

func (p *Producer) SendMessage(topic string, partition *int, headers map[string]string, bs []byte) error {
	var part int32 = -1
	if nil != partition {
		part = int32(*partition)
	}
	var hs []sarama.RecordHeader
	for k, v := range headers {
		hs = append(hs, sarama.RecordHeader{
			Key:   []byte(k),
			Value: []byte(v),
		})
	}
	message := &sarama.ProducerMessage{
		Topic:     topic,
		Partition: part,
		Value:     sarama.StringEncoder(bs),
		Headers:   hs,
		Timestamp: time.Now(),
	}
	var producer sarama.SyncProducer
	select {
	case <-p.done:
		return fmt.Errorf("producer closed")
	case producer = <-p.pool:
	}
	defer func() {
		p.pool <- producer
	}()
	//start := time.Now()
	part, _, err := producer.SendMessage(message)
	//TODO: log partition, offset etc.
	if nil != err {
		return err
	}
	// override log values if any
	//elapsed := time.Since(start)
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
	buf, err := json.Marshal(e.Payload())
	if err != nil {
		s.logger.Error().Str("op", "kafka.Send").Msg("failed to marshal message: " + err.Error())
		e.Nack(err)
		return
	}
	//TODO: support headers
	s.producer.SendMessage(s.config.Topic, s.config.Partition, nil, buf)
	s.logger.Info().Str("op", "kafka.Send").Msg("sent message on kafka topic")
	s.Lock()
	s.count++
	s.Unlock()
	e.Ack()
}
