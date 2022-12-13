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
	"github.com/boriwo/deepcopy"
	"github.com/rs/zerolog/log"
	"github.com/xmidt-org/ears/pkg/event"
	"github.com/xmidt-org/ears/pkg/filter"
	"github.com/xmidt-org/ears/pkg/secret"
	"github.com/xmidt-org/ears/pkg/tenant"
	"go.opentelemetry.io/otel/trace"
	"io/ioutil"
	"net"
	"net/http"
	"strings"
	"time"
)

var (
	tr     *http.Transport
	client *http.Client
)

const (
	HTTP_AUTH_TYPE_BASIC  = "basic"
	HTTP_AUTH_TYPE_SAT    = "sat"
	HTTP_AUTH_TYPE_OAUTH  = "oauth"
	HTTP_AUTH_TYPE_OAUTH2 = "oauth2"
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
		config:  *cfg,
		name:    name,
		plugin:  plugin,
		tid:     tid,
		secrets: secrets,
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
	if *f.config.EmptyPathValueRequired {
		pv, _, _ := evt.GetPathValue(f.config.ToPath)
		if pv != nil {
			return []event.Event{evt}
		}
	}
	// execute http request
	res, _, err := f.hitEndpoint(evt.Context(), evt)
	if err != nil {
		// legitimate filter nack because it involves an external service
		evt.Nack(err)
		return []event.Event{}
	}
	var resObj interface{}
	err = json.Unmarshal([]byte(res), &resObj)
	if err != nil {
		log.Ctx(evt.Context()).Error().Str("op", "filter").Str("filterType", "ws").Str("name", f.Name()).Msg(err.Error())
		if span := trace.SpanFromContext(evt.Context()); span != nil {
			span.AddEvent(err.Error())
		}
		evt.Ack()
		return []event.Event{}
	}
	if f.config.FromPath != "" {
		resEvt, _ := event.New(evt.Context(), resObj)
		resObj, _, _ = resEvt.GetPathValue(f.config.FromPath)
	}
	err = evt.DeepCopy()
	if err != nil {
		log.Ctx(evt.Context()).Error().Str("op", "filter").Str("filterType", "ws").Str("name", f.Name()).Msg(err.Error())
		if span := trace.SpanFromContext(evt.Context()); span != nil {
			span.AddEvent(err.Error())
		}
		evt.Ack()
		return []event.Event{}
	}
	_, _, err = evt.SetPathValue(f.config.ToPath, resObj, true)
	if err != nil {
		log.Ctx(evt.Context()).Error().Str("op", "filter").Str("filterType", "ws").Str("name", f.Name()).Msg(err.Error())
		if span := trace.SpanFromContext(evt.Context()); span != nil {
			span.AddEvent(err.Error())
		}
		evt.Ack()
		return []event.Event{}
	}
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

func (f *Filter) evalStr(evt event.Event, tt string) string {
	for {
		si := strings.Index(tt, "{")
		ei := strings.Index(tt, "}")
		if si < 0 || ei < 0 {
			break
		}
		path := tt[si+1 : ei]
		v, _, _ := evt.GetPathValue(path)
		v = deepcopy.DeepCopy(v)
		if !(si == 0 && ei == len(tt)-1) {
			switch vt := v.(type) {
			case string:
				tt = tt[0:si] + vt + tt[ei+1:]
			default:
				sv, _ := json.Marshal(vt)
				tt = tt[0:si] + string(sv) + tt[ei+1:]
			}
		} else {
			switch vt := v.(type) {
			case string:
				return vt
			default:
				sv, _ := json.Marshal(vt)
				return string(sv)
			}
		}
	}
	return tt
}

func (f *Filter) hitEndpoint(ctx context.Context, evt event.Event) (string, int, error) {
	e1 := f.evalStr(evt, f.config.Url)
	e2 := f.evalStr(evt, f.config.UrlPath)
	url := f.secrets.Secret(e1)
	if url == "" {
		url = e1
	}
	url = url + e2
	payload := f.evalStr(evt, f.config.Body)
	verb := f.evalStr(evt, f.config.Method)
	headers := f.config.Headers
	req, err := http.NewRequestWithContext(ctx, verb, url, bytes.NewBuffer([]byte(payload)))
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
	if f.config.Auth.Type == HTTP_AUTH_TYPE_BASIC {
		password := f.secrets.Secret(f.config.Auth.Password)
		if password == "" {
			password = f.config.Auth.Password
		}
		req.SetBasicAuth(f.config.Auth.Username, password)
	} else if f.config.Auth.Type == HTTP_AUTH_TYPE_SAT {
		return "", 0, errors.New("sat auth not supported")
	} else if f.config.Auth.Type == HTTP_AUTH_TYPE_OAUTH {
		return "", 0, errors.New("oauth auth not supported")
	} else if f.config.Auth.Type == HTTP_AUTH_TYPE_OAUTH2 {
		return "", 0, errors.New("oauth2 auth not supported")
	} else if f.config.Auth.Type != "" {
		return "", 0, errors.New("invalid auth type " + f.config.Auth.Type)
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
