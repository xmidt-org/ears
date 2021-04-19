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

//TODO: worker pool

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"github.com/Shopify/sarama"
	"github.com/rs/zerolog"
	"os"
	"time"

	"github.com/goccy/go-yaml"
	"github.com/xmidt-org/ears/pkg/event"
	pkgplugin "github.com/xmidt-org/ears/pkg/plugin"
	"github.com/xmidt-org/ears/pkg/receiver"
)

func NewReceiver(config interface{}) (receiver.Receiver, error) {
	var cfg ReceiverConfig
	var err error
	switch c := config.(type) {
	case string:
		err = yaml.Unmarshal([]byte(c), &cfg)
	case []byte:
		err = yaml.Unmarshal(c, &cfg)
	case ReceiverConfig:
		cfg = c
	case *ReceiverConfig:
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
	r := &Receiver{
		config: cfg,
		logger: logger,
	}
	return r, nil
}

////

// Setup is run at the beginning of a new session, before ConsumeClaim
func (r *Receiver) Setup(session sarama.ConsumerGroupSession) error {
	r.Lock()
	r.ConsumerGroupSession = session
	r.Unlock()
	// Mark the consumer as ready
	close(r.ready)
	return nil
}

// Cleanup is run at the end of a session, once all ConsumeClaim goroutines have exited
func (r *Receiver) Cleanup(session sarama.ConsumerGroupSession) error {
	return nil
}

// ConsumeClaim must start a consumer loop of ConsumerGroupClaim's Messages().
func (r *Receiver) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	// NOTE:
	// Do not move the code below to a goroutine.
	// The `ConsumeClaim` itself is called within a goroutine, see:
	// https://github.com/Shopify/sarama/blob/master/consumer_group.go#L27-L29
	for message := range claim.Messages() {
		//		log.Printf("Message claimed: value = %s, timestamp = %v, topic = %s", string(message.Value), message.Timestamp, message.Topic)
		if r.handler(message) {
			session.MarkMessage(message, "")
		}
	}
	return nil
}

func (r *Receiver) Start(ctx context.Context, groupHandler sarama.ConsumerGroupHandler, handler func(*sarama.ConsumerMessage) bool) {
	r.handler = handler
	r.wg.Add(1)
	defer r.wg.Done()
	for {
		// `Consume` should be called inside an infinite loop, when a
		// server-side rebalance happens, the consumer session will need to be
		// recreated to get the new claims
		if err := r.client.Consume(r.ctx, r.topics, groupHandler); err != nil {
			//ctx.Logger().Error("op", "Consumer.Start", "error", err)
		}
		// check if context was cancelled, signaling that the consumer should stop
		if nil != r.ctx.Err() {
			return
		}
		r.ready = make(chan bool)
	}
}

func (r *Receiver) Close(ctx context.Context) {
	<-r.ready // Await till the consumer has been set up
	r.cancel()
	r.wg.Wait()
	if err := r.client.Close(); err != nil {
		//ctx.Logger().Error("op", "Consumer.Close", "error", err)
	}
}

////

func (r *Receiver) getSaramaConfig(commitIntervalSec int) (*sarama.Config, error) {
	// init (custom) config, enable errors and notifications
	config := sarama.NewConfig()
	config.Consumer.Return.Errors = true
	config.Consumer.Offsets.Initial = sarama.OffsetNewest
	config.Consumer.Offsets.AutoCommit.Interval = time.Duration(commitIntervalSec) * time.Second
	// config.Group.Return.Notifications = true
	// if conf.ConsumeByPartitions {
	// 	config.Group.Mode = cluster.ConsumerModePartitions
	// }
	if "" != r.config.Version {
		v, err := sarama.ParseKafkaVersion(r.config.Version)
		if nil != err {
			return nil, err
		}
		config.Version = v
	}
	if 0 < r.config.ChannelBufferSize {
		config.ChannelBufferSize = r.config.ChannelBufferSize
	}
	config.Net.TLS.Enable = r.config.TLSEnable
	if "" != r.config.Username {
		config.Net.TLS.Enable = true
		config.Net.SASL.Enable = true
		config.Net.SASL.User = r.config.Username
		config.Net.SASL.Password = r.config.Password
	} else if "" != r.config.AccessCert {
		keypair, err := tls.X509KeyPair([]byte(r.config.AccessCert), []byte(r.config.AccessKey))
		if err != nil {
			return nil, err
		}
		caAuthorityPool := x509.NewCertPool()
		caAuthorityPool.AppendCertsFromPEM([]byte(r.config.CACert))
		tlsConfig := &tls.Config{
			Certificates: []tls.Certificate{keypair},
			RootCAs:      caAuthorityPool,
		}
		config.Net.TLS.Enable = true
		config.Net.TLS.Config = tlsConfig
	}
	return config, nil
}

////

func (r *Receiver) Receive(next receiver.NextFn) error {
	if r == nil {
		return &pkgplugin.Error{
			Err: fmt.Errorf("Receive called on <nil> pointer"),
		}
	}
	if next == nil {
		return &receiver.InvalidConfigError{
			Err: fmt.Errorf("next cannot be nil"),
		}
	}
	r.Lock()
	r.startTime = time.Now()
	r.next = next
	r.done = make(chan struct{})
	r.Unlock()
	go func() {
		//TODO: receive messages here
	}()
	r.logger.Info().Str("op", "kafka.Receive").Msg("waiting for receive done")
	<-r.done
	r.Lock()
	elapsedMs := time.Now().Sub(r.startTime).Milliseconds()
	throughput := 1000 * r.count / int(elapsedMs)
	cnt := r.count
	r.Unlock()
	r.logger.Info().Str("op", "kafka.Receive").Int("elapsedMs", int(elapsedMs)).Int("count", cnt).Int("throughput", throughput).Msg("receive done")
	return nil
}

func (r *Receiver) Count() int {
	r.Lock()
	defer r.Unlock()
	return r.count
}

func (r *Receiver) StopReceiving(ctx context.Context) error {
	r.Lock()
	if r.done != nil {
		r.done <- struct{}{}
	}
	r.Unlock()
	return nil
}

func (r *Receiver) Trigger(e event.Event) {
	r.Lock()
	next := r.next
	r.Unlock()
	next(e)
}
