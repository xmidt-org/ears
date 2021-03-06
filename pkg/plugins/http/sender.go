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
	"bytes"
	"context"
	"encoding/json"
	"github.com/goccy/go-yaml"
	"github.com/xmidt-org/ears/internal/pkg/rtsemconv"
	"github.com/xmidt-org/ears/pkg/event"
	pkgplugin "github.com/xmidt-org/ears/pkg/plugin"
	"github.com/xmidt-org/ears/pkg/secret"
	"github.com/xmidt-org/ears/pkg/sender"
	"github.com/xmidt-org/ears/pkg/tenant"
	"go.opentelemetry.io/otel"
	"io"
	"io/ioutil"
	"net/http"
	"time"
)

const DEFAULT_TIMEOUT = 10

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
	err = cfg.Validate()
	if err != nil {
		return nil, err
	}
	//TODO Does this live here?
	//TODO Make this a configuration?
	http.DefaultTransport.(*http.Transport).MaxIdleConnsPerHost = 100
	return &Sender{
		client: &http.Client{
			Timeout: DEFAULT_TIMEOUT * time.Second,
		},
		config: cfg,
		name:   name,
		plugin: plugin,
		tid:    tid,
	}, nil
}

func (s *Sender) SetTraceId(r *http.Request, traceId string) {
	r.Header.Set("traceId", traceId)
}

func (s *Sender) Send(event event.Event) {
	if event.Trace() {
		tracer := otel.Tracer(rtsemconv.EARSTracerName)
		_, span := tracer.Start(event.Context(), "httpSender")
		defer span.End()
	}
	payload := event.Payload()
	body, err := json.Marshal(payload)
	if err != nil {
		event.Nack(err)
		return
	}
	req, err := http.NewRequest(s.config.Method, s.config.Url, bytes.NewReader(body))
	if err != nil {
		event.Nack(err)
		return
	}
	ctx := event.Context()
	traceId := ctx.Value("traceId")
	if traceId != nil {
		s.SetTraceId(req, traceId.(string))
	}
	resp, err := s.client.Do(req)
	if err != nil {
		event.Nack(err)
		return
	}
	io.Copy(ioutil.Discard, resp.Body)
	defer resp.Body.Close()
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusBadRequest {
		event.Nack(&BadHttpStatusError{resp.StatusCode})
		return
	}
	event.Ack()
}

func (s *Sender) Unwrap() sender.Sender {
	return nil
}

func (s *Sender) StopSending(ctx context.Context) {
	//nothing
}

func (r *Sender) Config() interface{} {
	return r.config
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
