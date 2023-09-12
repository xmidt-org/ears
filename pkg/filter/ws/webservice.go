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
	"github.com/xmidt-org/ears/internal/pkg/syncer"
	"github.com/xmidt-org/ears/pkg/event"
	"github.com/xmidt-org/ears/pkg/filter"
	"github.com/xmidt-org/ears/pkg/hasher"
	"github.com/xmidt-org/ears/pkg/secret"
	"github.com/xmidt-org/ears/pkg/tenant"
	"go.opentelemetry.io/otel/trace"
	"golang.org/x/oauth2/clientcredentials"
	"io/ioutil"
	"net"
	"net/http"
	"strings"
	"time"
)

const (
	HTTP_AUTH_TYPE_BASIC  = "basic"
	HTTP_AUTH_TYPE_SAT    = "sat"
	HTTP_AUTH_TYPE_OAUTH  = "oauth"
	HTTP_AUTH_TYPE_OAUTH2 = "oauth2"

	SAT_URL = "https://sat-prod.codebig2.net/oauth/token"
)

type (
	SatToken struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
		Scope       string `json:"scope"`
		TokenType   string `json:"token_type"`
		ExpiresAt   int64  `json:"timestamp"`
	}
)

func NewFilter(tid tenant.Id, plugin string, name string, config interface{}, secrets secret.Vault, tableSyncer syncer.DeltaSyncer) (*Filter, error) {
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
		config:    *cfg,
		name:      name,
		plugin:    plugin,
		tid:       tid,
		secrets:   secrets,
		clients:   make(map[string]*http.Client),
		satTokens: make(map[string]*SatToken),
	}
	f.MetricFilter = filter.NewMetricFilter(tableSyncer)
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
		f.LogError()
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
		f.LogError()
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
		f.LogError()
		return []event.Event{}
	}
	_, _, err = evt.SetPathValue(f.config.ToPath, resObj, true)
	if err != nil {
		log.Ctx(evt.Context()).Error().Str("op", "filter").Str("filterType", "ws").Str("name", f.Name()).Msg(err.Error())
		if span := trace.SpanFromContext(evt.Context()); span != nil {
			span.AddEvent(err.Error())
		}
		evt.Ack()
		f.LogError()
		return []event.Event{}
	}
	log.Ctx(evt.Context()).Debug().Str("op", "filter").Str("filterType", "ws").Str("name", f.Name()).Msg("ws")
	f.LogSuccess()
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

func (f *Filter) getSatBearerToken(ctx context.Context, clientId string, clientSecret string) string {
	//curl -s -X POST -H "X-Client-Id: ***" -H "X-Client-Secret: ***" -H "Cache-Control: no-cache" https://sat-prod.codebig2.net/oauth/token
	//echo "Bearer $TOKEN"
	f.Lock()
	if time.Now().Unix() >= f.satTokens[clientId].ExpiresAt {
		delete(f.satTokens, clientId)
	}
	f.Unlock()
	if f.satTokens[clientId].AccessToken != "" {
		f.Lock()
		token := f.satTokens[clientId].TokenType + " " + f.satTokens[clientId].AccessToken
		f.Unlock()
		return token
	}
	httpClient := &http.Client{
		Timeout: 10 * time.Second,
	}
	req, err := http.NewRequest("POST", SAT_URL, nil)
	if err != nil {
		return ""
	}
	req.Header.Add("X-Client-Id", clientId)
	req.Header.Add("X-Client-Secret", clientSecret)
	req.Header.Add("Cache-Control", "no-cache")
	resp, err := httpClient.Do(req)
	if err != nil {
		return ""
	}
	buf, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()
	var satToken SatToken
	err = json.Unmarshal(buf, &satToken)
	if err != nil {
		return ""
	}
	satToken.ExpiresAt = time.Now().Unix() + int64(satToken.ExpiresIn)
	f.Lock()
	f.satTokens[clientId] = &satToken
	f.Unlock()
	return satToken.TokenType + " " + satToken.AccessToken
}

func (f *Filter) hitEndpoint(ctx context.Context, evt event.Event) (string, int, error) {
	e1 := f.evalStr(evt, f.config.Url)
	e2 := f.evalStr(evt, f.config.UrlPath)
	url := f.secrets.Secret(ctx, e1)
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
	var client *http.Client
	if f.config.Auth.Type == HTTP_AUTH_TYPE_BASIC {
		password := f.secrets.Secret(ctx, f.config.Auth.Password)
		if password == "" {
			password = f.config.Auth.Password
		}
		req.SetBasicAuth(f.config.Auth.Username, password)
		var ok bool
		f.RLock()
		client, ok = f.clients[HTTP_AUTH_TYPE_BASIC]
		f.RUnlock()
		if !ok {
			client = InitHttpTransportWithDialer()
			f.Lock()
			f.clients[HTTP_AUTH_TYPE_BASIC] = client
			f.Unlock()
		}
	} else if f.config.Auth.Type == HTTP_AUTH_TYPE_SAT {
		var ok bool
		f.RLock()
		client, ok = f.clients[HTTP_AUTH_TYPE_BASIC]
		f.RUnlock()
		if !ok {
			client = InitHttpTransportWithDialer()
			f.Lock()
			f.clients[HTTP_AUTH_TYPE_BASIC] = client
			f.Unlock()
		}
		token := f.getSatBearerToken(ctx, f.secrets.Secret(ctx, f.config.Auth.ClientID), f.secrets.Secret(ctx, f.config.Auth.ClientSecret))
		req.Header.Add("Authorization", token)
	} else if f.config.Auth.Type == HTTP_AUTH_TYPE_OAUTH {
		return "", 0, errors.New("oauth auth not supported")
	} else if f.config.Auth.Type == HTTP_AUTH_TYPE_OAUTH2 {
		var ok bool
		f.RLock()
		client, ok = f.clients[HTTP_AUTH_TYPE_OAUTH2+"-"+url]
		f.RUnlock()
		if !ok {
			conf := &clientcredentials.Config{
				ClientID:     f.secrets.Secret(ctx, f.config.Auth.ClientID),
				ClientSecret: f.secrets.Secret(ctx, f.config.Auth.ClientSecret),
				TokenURL:     f.secrets.Secret(ctx, f.config.Auth.TokenURL),
				Scopes:       f.config.Auth.Scopes,
			}
			client = conf.Client(context.Background())
			f.Lock()
			f.clients[HTTP_AUTH_TYPE_OAUTH2+"-"+url] = client
			f.Unlock()
		}
	} else if f.config.Auth.Type == "" {
		var ok bool
		f.RLock()
		client, ok = f.clients[HTTP_AUTH_TYPE_BASIC]
		f.RUnlock()
		if !ok {
			client = InitHttpTransportWithDialer()
			f.Lock()
			f.clients[HTTP_AUTH_TYPE_BASIC] = client
			f.Unlock()
		}
	} else {
		return "", 0, errors.New("invalid auth type " + f.config.Auth.Type)
	}
	if client == nil {
		return "", 0, errors.New("no client found for auth type " + f.config.Auth.Type)
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

func InitHttpTransportWithDialer() *http.Client {
	var dialer net.Dialer
	tr := &http.Transport{
		MaxIdleConnsPerHost:   100,
		ResponseHeaderTimeout: 1000 * time.Millisecond,
		Proxy:                 http.ProxyFromEnvironment,
		Dial:                  dialer.Dial,
	}
	go func() {
		for {
			time.Sleep(10 * time.Second)
			tr.CloseIdleConnections()
		}
	}()
	client := &http.Client{Transport: tr}
	client.Timeout = 3000 * time.Millisecond
	return client
}

func (f *Filter) Hash() string {
	cfg := ""
	if f.Config() != nil {
		buf, _ := json.Marshal(f.Config())
		if buf != nil {
			cfg = string(buf)
		}
	}
	str := f.name + f.plugin + cfg
	hash := hasher.String(str)
	return hash
}
