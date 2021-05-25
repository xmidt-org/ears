package http

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/goccy/go-yaml"
	"github.com/xmidt-org/ears/pkg/event"
	pkgplugin "github.com/xmidt-org/ears/pkg/plugin"
	"github.com/xmidt-org/ears/pkg/sender"
	"io"
	"io/ioutil"
	"net/http"
	"time"
)

const DEFAULT_TIMEOUT = 10

func NewSender(config interface{}) (sender.Sender, error) {
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
	return &Sender{
		client: &http.Client{
			Timeout: DEFAULT_TIMEOUT * time.Second,
		},
		method: cfg.Method,
		url:    cfg.Url,
	}, nil
}

func (s *Sender) SetTraceId(r *http.Request, traceId string) {
	r.Header.Set("traceId", traceId)
}

func (s *Sender) Send(event event.Event) {
	payload := event.Payload()
	body, err := json.Marshal(payload)
	if err != nil {
		event.Nack(err)
		return
	}

	req, err := http.NewRequest(s.method, s.url, bytes.NewReader(body))
	if err != nil {
		event.Nack(err)
	}

	ctx := event.Context()
	traceId := ctx.Value("traceId")
	if traceId != nil {
		s.SetTraceId(req, traceId.(string))
	}

	resp, err := s.client.Do(req)
	if err != nil {
		event.Nack(err)
	}

	io.Copy(ioutil.Discard, resp.Body)
	defer resp.Body.Close()
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusBadRequest {
		event.Nack(&BadHttpStatusError{resp.StatusCode})
	}
}

func (s *Sender) Unwrap() sender.Sender {
	return nil
}

func (s *Sender) StopSending(ctx context.Context) {
	//nothing
}
