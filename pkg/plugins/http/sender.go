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
	"errors"
	"github.com/boriwo/deepcopy"
	"github.com/goccy/go-yaml"
	"github.com/rs/zerolog/log"
	"github.com/xmidt-org/ears/internal/pkg/rtsemconv"
	"github.com/xmidt-org/ears/internal/pkg/syncer"
	"github.com/xmidt-org/ears/pkg/app"
	"github.com/xmidt-org/ears/pkg/event"
	"github.com/xmidt-org/ears/pkg/hasher"
	pkgplugin "github.com/xmidt-org/ears/pkg/plugin"
	"github.com/xmidt-org/ears/pkg/secret"
	"github.com/xmidt-org/ears/pkg/sender"
	"github.com/xmidt-org/ears/pkg/tenant"
	"go.opentelemetry.io/contrib/propagators/b3"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/global"
	"go.opentelemetry.io/otel/metric/unit"
	"go.opentelemetry.io/otel/trace"
	"golang.org/x/oauth2/clientcredentials"
	"io"
	"net"
	"net/http"
	nurl "net/url"
	"os"
	"strings"
	"time"
)

const DEFAULT_TIMEOUT = 10

func NewSender(tid tenant.Id, plugin string, name string, config interface{}, secrets secret.Vault, tableSyncer syncer.DeltaSyncer) (sender.Sender, error) {
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
	cfg = cfg.WithDefaults()
	err = cfg.Validate()
	if err != nil {
		return nil, err
	}
	s := &Sender{
		client: &http.Client{
			Timeout: DEFAULT_TIMEOUT * time.Second,
		},
		config:    cfg,
		name:      name,
		plugin:    plugin,
		tid:       tid,
		secrets:   secrets,
		clients:   make(map[string]*http.Client),
		satTokens: make(map[string]*SatToken),
	}
	s.MetricPlugin = pkgplugin.NewMetricPlugin(tableSyncer, s.Hash)
	// metric recorders
	hostname, _ := os.Hostname()
	meter := global.Meter(rtsemconv.EARSMeterName)
	commonLabels := []attribute.KeyValue{
		attribute.String(rtsemconv.EARSPluginTypeLabel, rtsemconv.EARSPluginTypeHttpSender),
		attribute.String(rtsemconv.EARSPluginNameLabel, s.Name()),
		attribute.String(rtsemconv.EARSAppIdLabel, s.tid.AppId),
		attribute.String(rtsemconv.EARSOrgIdLabel, s.tid.OrgId),
		attribute.String(rtsemconv.HostnameLabel, hostname),
	}
	s.eventSuccessCounter = metric.Must(meter).
		NewInt64Counter(
			rtsemconv.EARSMetricEventSuccess,
			metric.WithDescription("measures the number of successful events"),
		).Bind(commonLabels...)
	s.eventFailureCounter = metric.Must(meter).
		NewInt64Counter(
			rtsemconv.EARSMetricEventFailure,
			metric.WithDescription("measures the number of unsuccessful events"),
		).Bind(commonLabels...)
	s.eventBytesCounter = metric.Must(meter).
		NewInt64Counter(
			rtsemconv.EARSMetricEventBytes,
			metric.WithDescription("measures the number of event bytes processed"),
			metric.WithUnit(unit.Bytes),
		).Bind(commonLabels...)
	s.eventProcessingTime = metric.Must(meter).
		NewInt64Histogram(
			rtsemconv.EARSMetricEventProcessingTime,
			metric.WithDescription("measures the time an event spends in ears"),
			metric.WithUnit(unit.Milliseconds),
		).Bind(commonLabels...)
	s.eventSendOutTime = metric.Must(meter).
		NewInt64Histogram(
			rtsemconv.EARSMetricEventSendOutTime,
			metric.WithDescription("measures the time ears spends to send an event to a downstream data sink"),
			metric.WithUnit(unit.Milliseconds),
		).Bind(commonLabels...)

	s.b3Propagator = b3.New()
	return s, nil
}

func (s *Sender) evalStr(evt event.Event, tt string) string {
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

func (s *Sender) getSatBearerToken(ctx context.Context, clientId string, clientSecret string) (string, error) {
	s.Lock()
	if time.Now().Unix() >= s.satTokens[clientId].ExpiresAt {
		delete(s.satTokens, clientId)
	}
	s.Unlock()
	if s.satTokens[clientId].AccessToken != "" {
		s.Lock()
		token := s.satTokens[clientId].TokenType + " " + s.satTokens[clientId].AccessToken
		s.Unlock()
		return token, nil
	}
	httpClient := &http.Client{
		Timeout: 10 * time.Second,
	}
	req, err := http.NewRequest("POST", s.secrets.Secret(ctx, "secret://SATUrl"), nil)
	if err != nil {
		return "", err
	}
	req.Header.Add("X-Client-Id", clientId)
	req.Header.Add("X-Client-Secret", clientSecret)
	req.Header.Add("Cache-Control", "no-cache")
	resp, err := httpClient.Do(req)
	if err != nil {
		return "", err
	}
	buf, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	var satToken SatToken
	err = json.Unmarshal(buf, &satToken)
	if err != nil {
		return "", err
	}
	satToken.ExpiresAt = time.Now().Unix() + int64(satToken.ExpiresIn)
	s.Lock()
	s.satTokens[clientId] = &satToken
	s.Unlock()
	return satToken.TokenType + " " + satToken.AccessToken, nil
}

func (s *Sender) hitEndpoint(ctx context.Context, evt event.Event) (string, int, error) {
	e1 := s.evalStr(evt, s.config.Url)
	e2 := s.evalStr(evt, s.config.UrlPath)
	url := s.secrets.Secret(ctx, e1)
	if url == "" {
		url = e1
	}
	url = url + e2
	payload := s.evalStr(evt, s.config.Body)
	verb := s.evalStr(evt, s.config.Method)
	headers := s.config.Headers
	req, err := http.NewRequestWithContext(ctx, verb, url, bytes.NewBuffer([]byte(payload)))
	if err != nil {
		return "", 0, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "ears")
	// add trace header to outbound call
	traceHeader := app.HeaderTraceId
	traceId := trace.SpanFromContext(ctx).SpanContext().TraceID().String()
	req.Header.Set(traceHeader, traceId)
	// add supplied headers
	for hk, hv := range headers {
		req.Header.Set(hk, hv)
	}
	var client *http.Client
	if s.config.Auth.Type == HTTP_AUTH_TYPE_BASIC {
		password := s.secrets.Secret(ctx, s.config.Auth.Password)
		if password == "" {
			password = s.config.Auth.Password
		}
		req.SetBasicAuth(s.config.Auth.Username, password)
		var ok bool
		s.RLock()
		client, ok = s.clients[HTTP_AUTH_TYPE_BASIC]
		s.RUnlock()
		if !ok {
			client = s.initHttpTransportWithDialer()
			s.Lock()
			s.clients[HTTP_AUTH_TYPE_BASIC] = client
			s.Unlock()
		}
	} else if s.config.Auth.Type == HTTP_AUTH_TYPE_SAT {
		var ok bool
		s.RLock()
		client, ok = s.clients[HTTP_AUTH_TYPE_BASIC]
		s.RUnlock()
		if !ok {
			client = s.initHttpTransportWithDialer()
			s.Lock()
			s.clients[HTTP_AUTH_TYPE_BASIC] = client
			s.Unlock()
		}
		token, err := s.getSatBearerToken(ctx, s.secrets.Secret(ctx, s.config.Auth.ClientID), s.secrets.Secret(ctx, s.config.Auth.ClientSecret))
		if err != nil {
			return "", 0, err
		}
		req.Header.Add("Authorization", token)
	} else if s.config.Auth.Type == HTTP_AUTH_TYPE_OAUTH {
		return "", 0, errors.New("oauth auth not supported")
	} else if s.config.Auth.Type == HTTP_AUTH_TYPE_OAUTH2 {
		var ok bool
		s.RLock()
		client, ok = s.clients[HTTP_AUTH_TYPE_OAUTH2+"-"+url]
		s.RUnlock()
		if !ok {
			conf := &clientcredentials.Config{
				ClientID:     s.secrets.Secret(ctx, s.config.Auth.ClientID),
				ClientSecret: s.secrets.Secret(ctx, s.config.Auth.ClientSecret),
				TokenURL:     s.secrets.Secret(ctx, s.config.Auth.TokenURL),
				Scopes:       s.config.Auth.Scopes,
			}
			if s.config.Auth.GrantType != "" {
				conf.EndpointParams = nurl.Values{"grant_type": []string{s.config.Auth.GrantType}}
			}
			client = conf.Client(context.Background())
			s.Lock()
			s.clients[HTTP_AUTH_TYPE_OAUTH2+"-"+url] = client
			s.Unlock()
		}
	} else if s.config.Auth.Type == "" {
		var ok bool
		s.RLock()
		client, ok = s.clients[HTTP_AUTH_TYPE_BASIC]
		s.RUnlock()
		if !ok {
			client = s.initHttpTransportWithDialer()
			s.Lock()
			s.clients[HTTP_AUTH_TYPE_BASIC] = client
			s.Unlock()
		}
	} else {
		return "", 0, errors.New("invalid auth type " + s.config.Auth.Type)
	}
	if client == nil {
		return "", 0, errors.New("no client found for auth type " + s.config.Auth.Type)
	}
	// send request
	start := time.Now()
	resp, err := client.Do(req)
	s.eventSendOutTime.Record(evt.Context(), time.Since(start).Milliseconds())
	s.eventBytesCounter.Add(evt.Context(), int64(len(payload)))
	if err != nil {
		log.Ctx(evt.Context()).Error().Str("op", "filter").Str("filterType", "ws").Str("name", s.Name()).Msg("Request error: " + err.Error())
		return "", 0, err
	}
	// read response
	var body []byte
	if resp != nil && resp.Body != nil {
		var readErr error
		body, readErr = io.ReadAll(resp.Body)
		if readErr != nil {
			log.Ctx(evt.Context()).Error().Str("op", "filter").Str("filterType", "ws").Str("name", s.Name()).Msg("Read error: " + readErr.Error())
			return "", resp.StatusCode, readErr
		}
		resp.Body.Close()
		if body == nil {
			return "", resp.StatusCode, nil
		}
	}
	return string(body), resp.StatusCode, nil
}

func (s *Sender) initHttpTransportWithDialer() *http.Client {
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

func (s *Sender) Send(event event.Event) {
	// execute http request
	res, _, err := s.hitEndpoint(event.Context(), event)
	if err != nil {
		log.Ctx(event.Context()).Error().Str("op", "filter").Str("filterType", "ws").Str("name", s.Name()).Msg(err.Error())
		// legitimate filter nack because it involves an external service
		event.Nack(err)
		s.eventFailureCounter.Add(event.Context(), 1)
		s.LogError()
		return
	}
	event.SetResponse(res)
	//req, err := http.NewRequest(s.config.Method, s.config.Url, bytes.NewReader(body))
	/*req, err := http.NewRequest(s.config.Method, s.config.Url, bytes.NewReader([]byte(body)))
	if err != nil {
		s.eventFailureCounter.Add(event.Context(), 1)
		s.LogError()
		event.Nack(err)
		return
	}
	ctx := event.Context()
	s.b3Propagator.Inject(ctx, propagation.HeaderCarrier(req.Header))
	start := time.Now()
	resp, err := s.client.Do(req)
	s.eventSendOutTime.Record(event.Context(), time.Since(start).Milliseconds())
	if err != nil {
		s.eventFailureCounter.Add(event.Context(), 1)
		s.LogError()
		event.Nack(err)
		return
	}
	buf, err := io.ReadAll(resp.Body)
	if err != nil {
		s.eventFailureCounter.Add(event.Context(), 1)
		s.LogError()
		event.Nack(err)
		return
	}
	event.SetResponse(string(buf))
	defer resp.Body.Close()
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusBadRequest {
		s.eventFailureCounter.Add(event.Context(), 1)
		s.LogError()
		event.Nack(&BadHttpStatusError{resp.StatusCode})
		return
	}*/
	s.eventSuccessCounter.Add(event.Context(), 1)
	s.LogSuccess()
	event.Ack()
}

func (s *Sender) Unwrap() sender.Sender {
	return nil
}

func (s *Sender) StopSending(ctx context.Context) {
	s.eventSuccessCounter.Unbind()
	s.eventFailureCounter.Unbind()
	s.eventBytesCounter.Unbind()
	s.eventProcessingTime.Unbind()
	s.eventSendOutTime.Unbind()
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

func (s *Sender) Hash() string {
	cfg := ""
	if s.Config() != nil {
		buf, _ := json.Marshal(s.Config())
		if buf != nil {
			cfg = string(buf)
		}
	}
	str := s.name + s.plugin + cfg
	hash := hasher.String(str)
	return hash
}
