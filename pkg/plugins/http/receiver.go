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
	"github.com/xmidt-org/ears/pkg/event"
	pkgplugin "github.com/xmidt-org/ears/pkg/plugin"
	"github.com/xmidt-org/ears/pkg/receiver"
	"github.com/xmidt-org/ears/pkg/secret"
	"github.com/xmidt-org/ears/pkg/tenant"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/global"
	"io/ioutil"
	"net/http"
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
	err = cfg.Validate()
	if err != nil {
		return nil, err
	}
	r := &Receiver{
		config: cfg,
		name:   name,
		plugin: plugin,
		tid:    tid,
		logger: event.GetEventLogger(),
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
		).Bind(commonLabels...)
	return r, nil
}

func (h *Receiver) GetTraceId(r *http.Request) string {
	return r.Header.Get("traceId")
}

func (h *Receiver) Receive(next receiver.NextFn) error {
	mux := http.NewServeMux()
	port := *h.config.Port
	h.logger.Info().Int("port", port).Msg("starting http receiver")
	h.srv = &http.Server{Addr: fmt.Sprintf(":%d", port), Handler: mux}
	mux.HandleFunc(h.config.Path, func(w http.ResponseWriter, r *http.Request) {
		defer fmt.Fprintln(w, "good")
		b, err := ioutil.ReadAll(r.Body)
		defer r.Body.Close()
		if err != nil {
			h.logger.Error().Str("error", err.Error()).Msg("error reading body")
			return
		}
		var body interface{}
		err = json.Unmarshal(b, &body)
		if err != nil {
			h.logger.Error().Str("error", err.Error()).Msg("error unmarshalling body")
			return
		}
		ctx := context.Background() // acknowledge timeout?
		h.eventBytesCounter.Add(ctx, int64(len(b)))
		event, err := event.New(ctx, body,
			event.WithAck(
				func(e event.Event) {
					h.eventSuccessCounter.Add(ctx, 1)
				}, func(e event.Event, err error) {
					log.Ctx(e.Context()).Error().Str("error", err.Error()).Msg("nack handling events")
					h.eventFailureCounter.Add(ctx, 1)
				},
			),
			event.WithTenant(h.Tenant()),
			event.WithSpan(h.Name()),
		)
		if err != nil {
			h.logger.Error().Str("error", err.Error()).Msg("error creating event")
		}
		traceId := h.GetTraceId(r)
		if traceId != "" {
			subCtx := context.WithValue(event.Context(), "traceId", traceId)
			event.SetContext(subCtx)
		}
		next(event)
	})
	return h.srv.ListenAndServe()
}

func (h *Receiver) StopReceiving(ctx context.Context) error {
	if h.srv != nil {
		h.logger.Info().Msg("shutting down http receiver")
		return h.srv.Shutdown(ctx)
	}
	return nil
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
