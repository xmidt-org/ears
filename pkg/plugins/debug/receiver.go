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

package debug

import (
	"context"
	"github.com/rs/zerolog"
	"github.com/xmidt-org/ears/internal/pkg/rtsemconv"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/global"
	"go.opentelemetry.io/otel/trace"
	"os"
	"sync"

	"fmt"
	"time"

	"github.com/goccy/go-yaml"
	"github.com/xmidt-org/ears/pkg/event"
	pkgplugin "github.com/xmidt-org/ears/pkg/plugin"
	"github.com/xmidt-org/ears/pkg/receiver"
)

//TODO a configuration option to make this configurable

var debugMaxTO = time.Second * 10 //Default acknowledge timeout (10 seconds)

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
	r.done = make(chan struct{})
	r.stopped = false
	r.next = next
	tracer := otel.Tracer(rtsemconv.EARSTracerName)
	r.Unlock()
	go func() {
		defer func() {
			r.Lock()
			if !r.stopped {
				r.done <- struct{}{}
			}
			r.Unlock()
		}()
		eventsDone := &sync.WaitGroup{}
		eventsDone.Add(*r.config.Rounds)
		for count := *r.config.Rounds; count != 0; {
			select {
			case <-r.done:
				return
			case <-time.After(time.Duration(*r.config.IntervalMs) * time.Millisecond):
				ctx, cancel := context.WithTimeout(context.Background(), debugMaxTO)
				var span trace.Span
				if *r.config.Trace {
					ctx, span = tracer.Start(ctx, "debugReceiver")
					span.SetAttributes(rtsemconv.EARSEventTrace)
				}
				e, err := event.New(ctx, r.config.Payload, event.WithAck(
					func(evt event.Event) {
						eventsDone.Done()
						if *r.config.Trace {
							span.AddEvent("ack")
							span.End()
						}
						r.eventSuccessRecorder.Add(ctx, 1.0)
						cancel()
					}, func(evt event.Event, err error) {
						r.logger.Error().Str("op", "debug.Receive").Msg("failed to process message: " + err.Error())
						eventsDone.Done()
						if *r.config.Trace {
							span.AddEvent("nack")
							span.RecordError(err)
							span.End()
						}
						r.eventFailureRecorder.Add(ctx, 1.0)
						cancel()
					}),
					event.WithTrace(*r.config.Trace))
				if err != nil {
					return
				}
				r.Trigger(e)
				if count > 0 {
					count--
				}
			}
		}
		eventsDone.Wait()
	}()
	<-r.done
	return nil
}

func (r *Receiver) StopReceiving(ctx context.Context) error {
	r.Lock()
	defer r.Unlock()
	if !r.stopped && r.done != nil {
		r.eventSuccessRecorder.Unbind()
		r.eventFailureRecorder.Unbind()
		close(r.done)
		r.stopped = true
	}
	return nil
}

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
	//zerolog.LevelFieldName = "log.level"
	r := &Receiver{
		config:  cfg,
		logger:  logger,
		stopped: true,
	}
	r.history = newHistory(*r.config.MaxHistory)
	// metric recorders
	meter := global.Meter(rtsemconv.EARSMeterName)
	commonLabels := []attribute.KeyValue{
		attribute.String("pluginType", "debug"),
	}
	r.eventSuccessRecorder = metric.Must(meter).
		NewFloat64Counter(
			rtsemconv.EARSMetricEventSuccess,
			metric.WithDescription("measures the number of successful debug events"),
		).Bind(commonLabels...)
	r.eventFailureRecorder = metric.Must(meter).
		NewFloat64Counter(
			rtsemconv.EARSMetricEventFailure,
			metric.WithDescription("measures the number of unsuccessful debug events"),
		).Bind(commonLabels...)
	return r, nil
}

func (r *Receiver) Count() int {
	return r.history.Count()
}

func (r *Receiver) Trigger(e event.Event) {
	// Ensure that `next` can be slow and locking here will not
	// prevent other requests from executing.
	r.Lock()
	next := r.next
	r.Unlock()
	r.history.Add(e)
	next(e)
}

func (r *Receiver) History() []event.Event {
	history := r.history.History()
	events := make([]event.Event, len(history))
	for i, h := range history {
		if e, ok := h.(event.Event); ok {
			events[i] = e
		}
	}
	return events
}

func (r *Receiver) Config() interface{} {
	return r.config
}

func (r *Receiver) Name() string {
	return ""
}

func (r *Receiver) Plugin() string {
	return "debug"
}
