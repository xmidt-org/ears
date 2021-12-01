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

package ws

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/rs/zerolog/log"
	"github.com/xmidt-org/ears/pkg/event"
	"github.com/xmidt-org/ears/pkg/filter"
	"github.com/xmidt-org/ears/pkg/secret"
	"github.com/xmidt-org/ears/pkg/tenant"
	"go.opentelemetry.io/otel/trace"
	"io/ioutil"
	"net"
	"net/http"
	"time"
)

var (
	tr     *http.Transport
	client *http.Client
)

func NewFilter(tid tenant.Id, plugin string, name string, config interface{}, secrets secret.Vault) (*Filter, error) {
	cfg, err := NewConfig(config)
	if err != nil {
		return nil, &filter.InvalidConfigError{
			Err: err,
		}
	}
	cfg = cfg.WithDefaults()
	err = cfg.Validate()
	if err != nil {
		return nil, err
	}
	f := &Filter{
		config: *cfg,
		name:   name,
		plugin: plugin,
		tid:    tid,
	}
	return f, nil
}

func (f *Filter) Filter(evt event.Event) []event.Event {
	if f == nil {
		evt.Nack(&filter.InvalidConfigError{
			Err: fmt.Errorf("<nil> pointer filter"),
		})
		return nil
	}
	// execute http request
	res, _, err := hitEndpoint(evt.Context(), f.config.Url, f.config.Body, f.config.Method, f.config.Headers, map[string]string{})
	if err != nil {
		evt.Nack(err)
		return []event.Event{}
	}
	var resObj interface{}
	err = json.Unmarshal([]byte(res), &resObj)
	if err != nil {
		evt.Nack(err)
		return []event.Event{}
	}
	obj, _, _ := evt.GetPathValue(f.config.FromPath)
	if obj == nil {
		evt.Nack(errors.New("nil object at path " + f.config.FromPath))
		return []event.Event{}
	}
	path := f.config.FromPath
	if f.config.ToPath != "" {
		path = f.config.ToPath
	}
	evt.SetPathValue(path, resObj, true)
	log.Ctx(evt.Context()).Debug().Str("op", "filter").Str("filterType", "ws").Str("name", f.Name()).Msg("ws")
	return []event.Event{evt}
}

func (f *Filter) Config() interface{} {
	if f == nil {
		return Config{}
	}
	return f.config
}

func (f *Filter) Name() string {
	return f.name
}

func (f *Filter) Plugin() string {
	return f.plugin
}

func (f *Filter) Tenant() tenant.Id {
	return f.tid
}

func hitEndpoint(ctx context.Context, url string, payload string, verb string, headers map[string]string, auth map[string]string) (string, int, error) {
	req, err := http.NewRequest(verb, url, bytes.NewBuffer([]byte(payload)))
	if err != nil {
		return "", 0, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "ears")
	// add trace header to outbound call
	traceHeader := "X-B3-TraceId"
	traceId := trace.SpanFromContext(ctx).SpanContext().TraceID().String()
	req.Header.Set(traceHeader, traceId)
	// add supplied headers
	for hk, hv := range headers {
		req.Header.Set(hk, hv)
	}
	if auth != nil {
		if auth["type"] == "basic" {
			req.SetBasicAuth(auth["username"], auth["password"])
		}
	}
	// send request
	resp, err := client.Do(req)
	if err != nil {
		return "", 0, err
	}
	// read response
	var body []byte
	if resp != nil && resp.Body != nil {
		var readErr error
		body, readErr = ioutil.ReadAll(resp.Body)
		if readErr != nil {
			return "", resp.StatusCode, readErr
		}
		resp.Body.Close()
		if body == nil {
			return "", resp.StatusCode, nil
		}
	}
	return string(body), resp.StatusCode, nil
}

func init() {
	var dialer net.Dialer
	InitHttpTransportWithDial(dialer.Dial)
}

func InitHttpTransportWithDial(dial func(network, addr string) (net.Conn, error)) {
	tr = &http.Transport{
		MaxIdleConnsPerHost:   100,
		ResponseHeaderTimeout: 1000 * time.Millisecond,
		Proxy:                 http.ProxyFromEnvironment,
		Dial:                  dial,
	}
	go func() {
		for {
			time.Sleep(10 * time.Second)
			tr.CloseIdleConnections()
		}
	}()
	client = &http.Client{Transport: tr}
	client.Timeout = 3000 * time.Millisecond
}
