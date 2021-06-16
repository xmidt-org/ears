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

package redis

import (
	"context"
	"encoding/json"
	"github.com/go-redis/redis"
	"github.com/goccy/go-yaml"
	"github.com/rs/zerolog"
	"github.com/xmidt-org/ears/internal/pkg/rtsemconv"
	"github.com/xmidt-org/ears/pkg/event"
	pkgplugin "github.com/xmidt-org/ears/pkg/plugin"
	"github.com/xmidt-org/ears/pkg/secret"
	"github.com/xmidt-org/ears/pkg/sender"
	"github.com/xmidt-org/ears/pkg/tenant"
	"go.opentelemetry.io/otel"
	"os"
)

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
	cfg = cfg.WithDefaults()
	err = cfg.Validate()
	if err != nil {
		return nil, err
	}
	logger := zerolog.New(os.Stdout).Level(zerolog.DebugLevel)
	//zerolog.LevelFieldName = "log.level"
	s := &Sender{
		name:   name,
		plugin: plugin,
		tid:    tid,
		config: cfg,
		logger: logger,
	}
	s.initPlugin()
	return s, nil
}

func (s *Sender) initPlugin() error {
	s.Lock()
	s.client = redis.NewClient(&redis.Options{
		Addr:     s.config.Endpoint,
		Password: "",
		DB:       0,
	})
	s.Unlock()
	return nil
}

func (s *Sender) Count() int {
	s.Lock()
	defer s.Unlock()
	return s.count
}

func (s *Sender) StopSending(ctx context.Context) {
	s.Lock()
	s.client.Close()
	s.Unlock()
}

func (s *Sender) Send(e event.Event) {
	if e.Trace() {
		tracer := otel.Tracer(rtsemconv.EARSTracerName)
		_, span := tracer.Start(e.Context(), "redisSender")
		defer span.End()
	}
	buf, err := json.Marshal(e.Payload())
	if err != nil {
		s.logger.Error().Str("op", "redis.Send").Msg("failed to marshal message: " + err.Error())
		e.Nack(err)
		return
	}
	err = s.client.Publish(s.config.Channel, string(buf)).Err()
	if err != nil {
		s.logger.Error().Str("op", "redis.Send").Msg("failed to send message on redis channel: " + err.Error())
		e.Nack(err)
		return
	}
	s.logger.Info().Str("op", "redis.Send").Msg("sent message on redis channel")
	s.Lock()
	s.count++
	s.Unlock()
	e.Ack()
}

func (s *Sender) Unwrap() sender.Sender {
	return s
}

func (s *Sender) Config() interface{} {
	return s.config
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
