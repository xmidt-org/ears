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

//TODO: test rebalancing
//TODO: implement retries

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Shopify/sarama"
	"github.com/rs/zerolog/log"
	"github.com/xmidt-org/ears/internal/pkg/rtsemconv"
	"github.com/xmidt-org/ears/pkg/panics"
	"github.com/xmidt-org/ears/pkg/secret"
	"github.com/xmidt-org/ears/pkg/tenant"
	"go.opentelemetry.io/contrib/instrumentation/github.com/Shopify/sarama/otelsarama"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/global"
	"go.opentelemetry.io/otel/metric/unit"
	"strings"
	"time"

	"github.com/goccy/go-yaml"
	"github.com/xmidt-org/ears/pkg/event"
	pkgplugin "github.com/xmidt-org/ears/pkg/plugin"
	"github.com/xmidt-org/ears/pkg/receiver"
)

func NewReceiver(tid tenant.Id, plugin string, name string, config interface{}, secrets secret.Vault) (receiver.Receiver, error) {
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
	ctx := context.Background()
	cctx, cancel := context.WithCancel(ctx)
	r := &Receiver{
		config:  cfg,
		name:    name,
		plugin:  plugin,
		tid:     tid,
		logger:  event.GetEventLogger(),
		cancel:  cancel,
		ctx:     cctx,
		ready:   make(chan bool),
		topics:  []string{cfg.Topic},
		stopped: true,
		secrets: secrets,
	}
	saramaConfig, err := r.getSaramaConfig(*r.config.CommitInterval)
	if err != nil {
		return nil, err
	}
	saramaChan := make(chan sarama.ConsumerGroup, 1)
	go func() {
		var client sarama.ConsumerGroup
		client, err = sarama.NewConsumerGroup(strings.Split(r.config.Brokers, ","), r.config.GroupId, saramaConfig)
		saramaChan <- client
	}()
	select {
	case client := <-saramaChan:
		r.client = client
	case <-time.After(5 * time.Second):
		return nil, errors.New("sarama timed out")
	}
	if nil != err {
		return nil, err
	}
	// metric recorders
	meter := global.Meter(rtsemconv.EARSMeterName)
	commonLabels := []attribute.KeyValue{
		attribute.String(rtsemconv.EARSPluginTypeLabel, rtsemconv.EARSPluginTypeKafkaReceiver),
		attribute.String(rtsemconv.EARSPluginNameLabel, r.Name()),
		attribute.String(rtsemconv.EARSAppIdLabel, r.tid.AppId),
		attribute.String(rtsemconv.EARSOrgIdLabel, r.tid.OrgId),
		attribute.String(rtsemconv.KafkaTopicLabel, r.config.Topic),
		attribute.String(rtsemconv.KafkaGroupIdLabel, r.config.GroupId),
		attribute.String(rtsemconv.EARSReceiverName, r.name),
	}
	r.eventSuccessCounter = metric.Must(meter).
		NewInt64Counter(
			rtsemconv.EARSMetricEventSuccess,
			metric.WithDescription("measures the number of successful events"),
		).Bind(commonLabels...)
	r.eventFailureCounter = metric.Must(meter).
		NewInt64Counter(
			rtsemconv.EARSMetricEventFailure,
			metric.WithDescription("measures the number of unsuccessful events"),
		).Bind(commonLabels...)
	r.eventBytesCounter = metric.Must(meter).
		NewInt64Counter(
			rtsemconv.EARSMetricEventBytes,
			metric.WithDescription("measures the number of event bytes processed"),
			metric.WithUnit(unit.Bytes),
		).Bind(commonLabels...)
	return r, nil
}

// Setup is run at the beginning of a new session, before ConsumeClaim
func (r *Receiver) Setup(session sarama.ConsumerGroupSession) error {
	r.Lock()
	r.ConsumerGroupSession = session
	r.Unlock()
	// mark the consumer as ready
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
		if r.handler(message) {
			session.MarkMessage(message, "")
		}
	}
	return nil
}

func (r *Receiver) Start(handler func(*sarama.ConsumerMessage) bool) {
	r.handler = handler
	r.wg.Add(1)
	defer r.wg.Done()
	for {
		// `Consume` should be called inside an infinite loop, when a
		// server-side rebalance happens, the consumer session will need to be
		// recreated to get the new claims
		//err := r.client.Consume(r.ctx, r.topics, otelsarama.WrapConsumerGroupHandler(r))
		err := r.client.Consume(r.ctx, r.topics, r)
		if err != nil { // the receiver itself is the group handler
			r.logger.Error().Str("op", "kafka.Start").Str("name", r.Name()).Str("tid", r.Tenant().ToString()).Msg(err.Error())
			//Sleep for a little bit to prevent busy loop
			time.Sleep(time.Second)
		} else {
			r.logger.Info().Str("op", "kafka.Start").Str("name", r.Name()).Str("tid", r.Tenant().ToString()).Msg("kafka consumer finished without error")
		}
		// check if context was canceled, signaling that the consumer should stop
		if nil != r.ctx.Err() {
			r.logger.Error().Str("op", "kafka.Start").Str("name", r.Name()).Str("tid", r.Tenant().ToString()).Msg("context canceled, stopping consumption")
			return
		}
		r.ready = make(chan bool)
	}
}

func (r *Receiver) Close() {
	//r.logger.Info().Str("op", "kafka.Close").Msg("starting tear down")
	<-r.ready // wait until consumer has been set up
	//r.logger.Info().Str("op", "kafka.Close").Msg("consumer ready")
	r.cancel()
	//r.logger.Info().Str("op", "kafka.Close").Msg("conext canceled")
	r.wg.Wait()
	//r.logger.Info().Str("op", "kafka.Close").Msg("wait group done")
	err := r.client.Close()
	if err != nil {
		r.logger.Error().Str("op", "kafka.Close").Str("name", r.Name()).Str("tid", r.Tenant().ToString()).Msg(err.Error())
	} else {
		r.logger.Info().Str("op", "kafka.Close").Str("name", r.Name()).Str("tid", r.Tenant().ToString()).Msg("kafka consumer closed")
	}
}

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
	if *r.config.ChannelBufferSize > 0 {
		config.ChannelBufferSize = *r.config.ChannelBufferSize
	}
	config.Net.TLS.Enable = r.config.TLSEnable
	if "" != r.config.Username {
		config.Net.TLS.Enable = true
		config.Net.SASL.Enable = true
		config.Net.SASL.User = r.config.Username
		config.Net.SASL.Password = r.secrets.Secret(r.config.Password)
		if config.Net.SASL.Password == "" {
			config.Net.SASL.Password = r.config.Password
		}

	} else if "" != r.config.AccessCert {
		accessCert := r.secrets.Secret(r.config.AccessCert)
		if accessCert == "" {
			accessCert = r.config.AccessCert
		}
		accessKey := r.secrets.Secret(r.config.AccessKey)
		if accessKey == "" {
			accessKey = r.config.AccessKey
		}
		caCert := r.secrets.Secret(r.config.CACert)
		if caCert == "" {
			caCert = r.config.CACert
		}
		keypair, err := tls.X509KeyPair([]byte(accessCert), []byte(accessKey))
		if err != nil {
			return nil, err
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
	return config, nil
}

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
	r.stopped = false
	r.Unlock()
	go func() {
		defer func() {
			p := recover()
			if p != nil {
				panicErr := panics.ToError(p)
				r.logger.Error().Str("op", "kafka.Receive").Str("error", panicErr.Error()).
					Str("stackTrace", panicErr.StackTrace()).Msg("A panic has occurred")
				r.Close()
			}
		}()
		r.Start(func(msg *sarama.ConsumerMessage) bool {
			defer func() {
				p := recover()
				if p != nil {
					panicErr := panics.ToError(p)
					r.logger.Error().Str("op", "kafka.Receive").Str("error", panicErr.Error()).
						Str("stackTrace", panicErr.StackTrace()).Msg("A panic has occurred while handling a message")
				}
			}()

			// bail if context has been canceled
			if r.ctx.Err() != nil {
				r.logger.Info().Str("op", "kafka.Receive").Str("name", r.Name()).Str("tid", r.Tenant().ToString()).Msg("abandoning message due to canceled context")
				return false
			}
			r.logger.Debug().Str("op", "kafka.Receive").Str("name", r.Name()).Str("tid", r.Tenant().ToString()).Str("topic", msg.Topic).Int("partition", int(msg.Partition)).Int("offset", int(msg.Offset)).Msg("message received")
			r.Lock()
			r.count++
			r.Unlock()
			var pl interface{}
			err := json.Unmarshal(msg.Value, &pl)
			if err != nil {
				r.logger.Error().Str("op", "kafka.Receive").Msg("cannot parse payload: " + err.Error())
				return false
			}
			ctx, cancel := context.WithTimeout(context.Background(), time.Duration(5)*time.Second)

			//extract otel tracing info
			ctx = otel.GetTextMapPropagator().Extract(ctx, otelsarama.NewConsumerMessageCarrier(msg))

			r.eventBytesCounter.Add(ctx, int64(len(msg.Value)))
			e, err := event.New(ctx, pl, event.WithAck(
				func(e event.Event) {
					log.Ctx(e.Context()).Debug().Str("op", "kafka.Receive").Str("name", r.Name()).Str("tid", r.Tenant().ToString()).Msg("processed message from kafka topic")
					r.eventSuccessCounter.Add(ctx, 1)
					cancel()
				},
				func(e event.Event, err error) {
					log.Ctx(e.Context()).Error().Str("op", "kafka.Receive").Str("name", r.Name()).Str("tid", r.Tenant().ToString()).Msg("failed to process message: " + err.Error())
					r.eventFailureCounter.Add(ctx, 1)
					cancel()
				}),
				event.WithOtelTracing(r.Name()),
				event.WithTenant(r.Tenant()),
				event.WithTracePayloadOnNack(*r.config.TracePayloadOnNack))
			if err != nil {
				r.logger.Error().Str("op", "kafka.Receive").Str("name", r.Name()).Str("tid", r.Tenant().ToString()).Msg("cannot create event: " + err.Error())
				return false
			}
			r.Trigger(e)
			return true
		})
	}()
	r.logger.Info().Str("op", "kafka.Receive").Str("name", r.Name()).Str("tid", r.Tenant().ToString()).Msg("waiting for receive done")
	<-r.done
	r.Lock()
	elapsedMs := time.Since(r.startTime).Milliseconds()
	throughput := 1000 * r.count / (int(elapsedMs) + 1)
	cnt := r.count
	r.Unlock()
	r.logger.Info().Str("op", "kafka.Receive").Str("name", r.Name()).Str("tid", r.Tenant().ToString()).Int("elapsedMs", int(elapsedMs)).Int("count", cnt).Int("throughput", throughput).Msg("receive done")
	return nil
}

func (r *Receiver) Count() int {
	r.Lock()
	defer r.Unlock()
	return r.count
}

func (r *Receiver) StopReceiving(ctx context.Context) error {
	r.logger.Info().Str("op", "kafka.StopReceiving").Str("name", r.Name()).Str("tid", r.Tenant().ToString()).Msg("stop receiving")
	r.Lock()
	if !r.stopped {
		r.stopped = true
		r.eventSuccessCounter.Unbind()
		r.eventFailureCounter.Unbind()
		r.eventBytesCounter.Unbind()
		close(r.done)
	}
	r.Unlock()
	r.logger.Info().Str("op", "kafka.StopReceiving").Str("name", r.Name()).Str("tid", r.Tenant().ToString()).Msg("done sent to receiver func")
	r.Close()
	r.logger.Info().Str("op", "kafka.StopReceiving").Str("name", r.Name()).Str("tid", r.Tenant().ToString()).Msg("kafka client closed")
	return nil
}

func (r *Receiver) Trigger(e event.Event) {
	r.Lock()
	next := r.next
	r.Unlock()
	next(e)
}

func (r *Receiver) Config() interface{} {
	return r.config
}

func (r *Receiver) Name() string {
	return r.name
}

func (r *Receiver) Plugin() string {
	return r.plugin
}

func (r *Receiver) Tenant() tenant.Id {
	return r.tid
}
