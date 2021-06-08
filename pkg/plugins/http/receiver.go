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
	"github.com/rs/zerolog"
	"github.com/xmidt-org/ears/pkg/event"
	pkgplugin "github.com/xmidt-org/ears/pkg/plugin"
	"github.com/xmidt-org/ears/pkg/receiver"
	"io/ioutil"
	"net/http"
	"os"
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
	err = cfg.Validate()
	if err != nil {
		return nil, err
	}

	logger := zerolog.New(os.Stdout).Level(zerolog.InfoLevel)
	return &Receiver{
		config: cfg,
		logger: &logger,
	}, nil
}

func (h *Receiver) GetTraceId(r *http.Request) string {
	return r.Header.Get("traceId")
}

func (h *Receiver) Receive(next receiver.NextFn) error {

	mux := http.NewServeMux()
	port := *h.config.Port
	h.logger.Info().Int("port", port).Msg("Starting http receiver")
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
		ctx := context.Background()
		event, err := event.New(ctx, body, event.WithAck(func(e event.Event) {
			//do nothing
		}, func(e event.Event, err error) {
			h.logger.Error().Str("error", err.Error()).Msg("Nack handling events")
		},
		))
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
		h.logger.Info().Msg("Shutting down HTTP receiver")
		return h.srv.Shutdown(ctx)
	}
	return nil
}

func (r *Receiver) Config() interface{} {
	return r.config
}

func (r *Receiver) Name() string {
	return ""
}

func (r *Receiver) Plugin() string {
	return "http"
}
