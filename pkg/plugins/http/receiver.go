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

package http

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/goccy/go-yaml"
	"github.com/rs/zerolog/log"
	"github.com/xmidt-org/ears/internal/pkg/rtsemconv"
	"github.com/xmidt-org/ears/internal/pkg/syncer"
	"github.com/xmidt-org/ears/pkg/event"
	"github.com/xmidt-org/ears/pkg/hasher"
	pkgplugin "github.com/xmidt-org/ears/pkg/plugin"
	"github.com/xmidt-org/ears/pkg/receiver"
	"github.com/xmidt-org/ears/pkg/secret"
	"github.com/xmidt-org/ears/pkg/tenant"
	"go.opentelemetry.io/contrib/propagators/b3"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/global"
	"go.opentelemetry.io/otel/metric/unit"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
	"io/ioutil"
	"net/http"
	"strings"
	"sync"
	"time"
)

func NewReceiver(tid tenant.Id, plugin string, name string, config interface{}, secrets secret.Vault, tableSyncer syncer.DeltaSyncer) (receiver.Receiver, error) {
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
		config:      cfg,
		name:        name,
		plugin:      plugin,
		tid:         tid,
		logger:      event.GetEventLogger(),
		currentSec:  time.Now().Unix(),
		tableSyncer: tableSyncer,
	}
	// metric recorders
	meter := global.Meter(rtsemconv.EARSMeterName)
	commonLabels := []attribute.KeyValue{
		attribute.String(rtsemconv.EARSPluginTypeLabel, rtsemconv.EARSPluginTypeHttpReceiver),
		attribute.String(rtsemconv.EARSPluginNameLabel, r.Name()),
		attribute.String(rtsemconv.EARSAppIdLabel, r.tid.AppId),
		attribute.String(rtsemconv.EARSOrgIdLabel, r.tid.OrgId),
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

func (r *Receiver) LogSuccess() {
	r.Lock()
	r.successCounter++
	if time.Now().Unix() != r.currentSec {
		r.successVelocityCounter = r.currentSuccessVelocityCounter
		r.currentSuccessVelocityCounter = 0
		r.currentSec = time.Now().Unix()
	}
	r.currentSuccessVelocityCounter++
	r.Unlock()
}

func (r *Receiver) logError() {
	r.Lock()
	r.errorCounter++
	if time.Now().Unix() != r.currentSec {
		r.errorVelocityCounter = r.currentErrorVelocityCounter
		r.currentErrorVelocityCounter = 0
		r.currentSec = time.Now().Unix()
	}
	r.currentErrorVelocityCounter++
	r.Unlock()
}

func (r *Receiver) Receive(next receiver.NextFn) error {
	r.next = next
	mux := http.NewServeMux()
	port := *r.config.Port
	r.logger.Info().Int("port", port).Str("path", r.config.Path).Msg("starting http receiver")
	r.srv = &http.Server{Addr: fmt.Sprintf(":%d", port), Handler: mux}
	b3Propagator := b3.New()

	mux.HandleFunc(r.config.Path, func(w http.ResponseWriter, req *http.Request) {
		if r.config.Method != "" && !strings.EqualFold(r.config.Method, strings.ToLower(req.Method)) {
			r.logger.Error().Str("method", req.Method).Msg("unexpected method error")
			return
		}
		b, err := ioutil.ReadAll(req.Body)
		defer req.Body.Close()
		if err != nil {
			r.logger.Error().Str("error", err.Error()).Msg("error reading body")
			return
		}
		var body interface{}
		err = json.Unmarshal(b, &body)
		if err != nil {
			r.logger.Error().Str("error", err.Error()).Msg("error unmarshalling body")
			return
		}
		ctx, cancel := context.WithTimeout(context.Background(), time.Duration(5)*time.Second)

		//extract any trace information
		ctx = b3Propagator.Extract(ctx, propagation.HeaderCarrier(req.Header))

		r.eventBytesCounter.Add(ctx, int64(len(b)))
		var wg sync.WaitGroup
		wg.Add(1)
		name := r.name
		if name == "" {
			name = "http"
		}
		metadata := map[string]interface{}{name: map[string]interface{}{
			"path":         req.URL.Path,
			"relativePath": req.URL.Path[len(r.config.Path):],
			"method":       req.Method,
		}}
		event, err := event.New(ctx, body,
			event.WithAck(
				func(e event.Event) {
					w.Header().Set("Content-Type", "application/json")
					w.Header().Set("User-Agent", "ears")
					w.WriteHeader(*r.config.SuccessStatus)
					resp := Response{
						Status: &Status{
							Code: *r.config.SuccessStatus,
						},
						Tracing: &Tracing{
							TraceId: trace.SpanFromContext(ctx).SpanContext().TraceID().String(),
						},
					}
					json.NewEncoder(w).Encode(resp)
					wg.Done()
					r.eventSuccessCounter.Add(ctx, 1)
					r.LogSuccess()
					cancel()
				}, func(e event.Event, err error) {
					w.Header().Set("Content-Type", "application/json")
					w.Header().Set("User-Agent", "ears")
					w.WriteHeader(*r.config.FailureStatus)
					resp := Response{
						Status: &Status{
							Code: *r.config.FailureStatus,
						},
						Tracing: &Tracing{
							TraceId: trace.SpanFromContext(ctx).SpanContext().TraceID().String(),
						},
					}
					json.NewEncoder(w).Encode(resp)
					wg.Done()
					log.Ctx(e.Context()).Error().Str("error", err.Error()).Msg("nack handling events")
					r.eventFailureCounter.Add(ctx, 1)
					r.logError()
					cancel()
				},
			),
			event.WithTenant(r.Tenant()),
			event.WithOtelTracing(r.Name()),
			event.WithTracePayloadOnNack(*r.config.TracePayloadOnNack),
			event.WithMetadata(metadata),
		)
		if err != nil {
			r.logger.Error().Str("error", err.Error()).Msg("error creating event")
		}
		next(event)
		wg.Wait()
	})
	return r.srv.ListenAndServe()
}

func (r *Receiver) StopReceiving(ctx context.Context) error {
	if r.srv != nil {
		r.logger.Info().Msg("shutting down http receiver")
		return r.srv.Shutdown(ctx)
	}
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

func (r *Receiver) getLocalMetric() *syncer.EarsMetric {
	r.Lock()
	defer r.Unlock()
	metrics := &syncer.EarsMetric{
		r.successCounter,
		r.errorCounter,
		0,
		r.successVelocityCounter,
		r.errorVelocityCounter,
		0,
		r.currentSec,
		0,
	}
	return metrics
}

func (r *Receiver) EventSuccessCount() int {
	hash := r.Hash()
	r.tableSyncer.WriteMetrics(hash, r.getLocalMetric())
	return r.tableSyncer.ReadMetrics(hash).SuccessCount
}

func (r *Receiver) EventSuccessVelocity() int {
	hash := r.Hash()
	r.tableSyncer.WriteMetrics(hash, r.getLocalMetric())
	return r.tableSyncer.ReadMetrics(hash).SuccessVelocity
}

func (r *Receiver) EventErrorCount() int {
	hash := r.Hash()
	r.tableSyncer.WriteMetrics(hash, r.getLocalMetric())
	return r.tableSyncer.ReadMetrics(hash).ErrorCount
}

func (r *Receiver) EventErrorVelocity() int {
	hash := r.Hash()
	r.tableSyncer.WriteMetrics(hash, r.getLocalMetric())
	return r.tableSyncer.ReadMetrics(hash).ErrorVelocity
}

func (r *Receiver) EventTs() int64 {
	hash := r.Hash()
	r.tableSyncer.WriteMetrics(hash, r.getLocalMetric())
	return r.tableSyncer.ReadMetrics(hash).LastEventTs
}

func (r *Receiver) Hash() string {
	cfg := ""
	if r.Config() != nil {
		buf, _ := json.Marshal(r.Config())
		if buf != nil {
			cfg = string(buf)
		}
	}
	str := r.name + r.plugin + cfg
	hash := hasher.String(str)
	return hash
}
