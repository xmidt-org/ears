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
	"fmt"
	"github.com/go-redis/redis"
	"github.com/rs/zerolog/log"
	"github.com/xmidt-org/ears/internal/pkg/rtsemconv"
	"github.com/xmidt-org/ears/pkg/secret"
	"github.com/xmidt-org/ears/pkg/tenant"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/global"
	"go.opentelemetry.io/otel/metric/unit"
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
	r := &Receiver{
		config:     cfg,
		name:       name,
		plugin:     plugin,
		tid:        tid,
		logger:     event.GetEventLogger(),
		stopped:    true,
		currentSec: time.Now().Second(),
	}
	// metric recorders
	meter := global.Meter(rtsemconv.EARSMeterName)
	commonLabels := []attribute.KeyValue{
		attribute.String(rtsemconv.EARSPluginTypeLabel, rtsemconv.EARSPluginTypeRedisReceiver),
		attribute.String(rtsemconv.EARSPluginNameLabel, r.Name()),
		attribute.String(rtsemconv.EARSAppIdLabel, r.tid.AppId),
		attribute.String(rtsemconv.EARSOrgIdLabel, r.tid.OrgId),
		attribute.String(rtsemconv.RedisChannelLabel, r.config.Channel),
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

func (r *Receiver) logSuccess() {
	r.Lock()
	r.successCounter++
	if time.Now().Second() != r.currentSec {
		r.successVelocityCounter = r.currentSuccessVelocityCounter
		r.currentSuccessVelocityCounter = 0
		r.currentSec = time.Now().Second()
	}
	r.currentSuccessVelocityCounter++
	r.Unlock()
}

func (r *Receiver) logError() {
	r.Lock()
	r.errorCounter++
	if time.Now().Second() != r.currentSec {
		r.errorVelocityCounter = r.currentErrorVelocityCounter
		r.currentErrorVelocityCounter = 0
		r.currentSec = time.Now().Second()
	}
	r.currentErrorVelocityCounter++
	r.Unlock()
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
			r.Lock()
			if !r.stopped {
				r.done <- struct{}{}
			}
			r.Unlock()
		}()
		r.redisClient = redis.NewClient(&redis.Options{
			Addr:     r.config.Endpoint,
			Password: "",
			DB:       0,
		})
		defer r.redisClient.Close()
		r.pubsub = r.redisClient.Subscribe(r.config.Channel)
		defer r.pubsub.Close()
		for {
			// could have a pool of go routines consuming from this channel here
			msg := <-r.pubsub.Channel()
			if r.stopped {
				r.logger.Info().Str("op", "redis.Receive").Msg("stopping receive loop")
				return
			}
			var pl interface{}
			err := json.Unmarshal([]byte(msg.Payload), &pl)
			if err != nil {
				r.logger.Error().Str("op", "redis.Receive").Msg("cannot parse payload: " + err.Error())
				continue
			}
			ctx, cancel := context.WithTimeout(context.Background(), time.Duration(5)*time.Second)
			r.eventBytesCounter.Add(ctx, int64(len(msg.Payload)))
			r.Lock()
			r.count++
			r.logger.Debug().Str("op", "redis.Receive").Str("name", r.Name()).Str("tid", r.Tenant().ToString()).Int("receiveCount", r.count).Msg("received message on redis channel")
			r.Unlock()
			// note: if we just pass msg.Payload into event, redis will blow up with an out of memory error within a
			// few seconds - possibly a bug in the client library
			e, err := event.New(ctx, pl, event.WithAck(
				func(e event.Event) {
					log.Ctx(e.Context()).Debug().Str("op", "redis.Receive").Str("name", r.Name()).Str("tid", r.Tenant().ToString()).Msg("processed message from redis channel")
					r.eventSuccessCounter.Add(ctx, 1)
					r.logSuccess()
					cancel()
				},
				func(e event.Event, err error) {
					log.Ctx(e.Context()).Error().Str("op", "redis.Receive").Msg("failed to process message: " + err.Error())
					r.eventFailureCounter.Add(ctx, 1)
					r.logError()
					cancel()
				}),
				event.WithOtelTracing(r.Name()),
				event.WithTenant(r.Tenant()),
				event.WithTracePayloadOnNack(*r.config.TracePayloadOnNack),
			)
			if err != nil {
				r.logger.Error().Str("op", "redis.Receive").Msg("cannot create event: " + err.Error())
				continue
			}
			r.Trigger(e)
		}
	}()
	r.logger.Info().Str("op", "redis.Receive").Msg("waiting for receive done")
	<-r.done
	r.Lock()
	elapsedMs := time.Since(r.startTime).Milliseconds()
	throughput := 1000 * r.count / (int(elapsedMs) + 1)
	cnt := r.count
	r.Unlock()
	r.logger.Info().Str("op", "redis.Receive").Int("elapsedMs", int(elapsedMs)).Int("count", cnt).Int("throughput", throughput).Msg("receive done")
	return nil
}

func (r *Receiver) Count() int {
	r.Lock()
	defer r.Unlock()
	return r.count
}

func (r *Receiver) StopReceiving(ctx context.Context) error {
	r.Lock()
	if !r.stopped {
		r.stopped = true
		r.pubsub.Unsubscribe(r.config.Channel)
		r.pubsub.Close()
		r.eventSuccessCounter.Unbind()
		r.eventFailureCounter.Unbind()
		r.eventBytesCounter.Unbind()
		close(r.done)
	}
	r.Unlock()
	return nil
}

func (r *Receiver) Trigger(e event.Event) {
	r.Lock()
	next := r.next
	r.Unlock()
	if next != nil {
		next(e)
	}
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

func (r *Receiver) EventSuccessCount() int {
	r.Lock()
	defer r.Unlock()
	return r.successCounter
}

func (r *Receiver) EventSuccessVelocity() int {
	r.Lock()
	defer r.Unlock()
	return r.successVelocityCounter
}

func (r *Receiver) EventErrorCount() int {
	r.Lock()
	defer r.Unlock()
	return r.errorCounter
}

func (r *Receiver) EventErrorVelocity() int {
	r.Lock()
	defer r.Unlock()
	return r.errorVelocityCounter
}
