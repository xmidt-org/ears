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

package debug

import (
	"context"
	"fmt"

	"github.com/goccy/go-yaml"
	"github.com/xmidt-org/ears/pkg/event"
	pkgplugin "github.com/xmidt-org/ears/pkg/plugin"
	"github.com/xmidt-org/ears/pkg/sender"
)

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

	cfg = cfg.WithDefaults()

	err = cfg.Validate()
	if err != nil {
		return nil, err
	}

	s := &Sender{
		config: cfg,
	}

	switch s.config.Destination {
	case DestinationDevNull:
		s.destination = nil
	case DestinationStdout:
		s.destination = &SendStdout{}
	case DestinationStderr:
		s.destination = &SendStderr{}
	case DestinationCustom:
		s.destination = s.config.Writer
	default:
		return nil, &pkgplugin.InvalidConfigError{
			Err: fmt.Errorf("config.Destination value invalid"),
		}
	}

	s.history = newHistory(*s.config.MaxHistory)

	return s, nil

}

func (s *Sender) Send(ctx context.Context, e event.Event) error {
	s.history.Add(e)
	fmt.Printf("SEND\n")

	if s.destination != nil {
		return s.destination.Write(e)
	}

	return nil
}

func (s *Sender) Unwrap() sender.Sender {
	return s
}

func (s *Sender) Count() int {
	return s.history.Count()
}

func (s *Sender) History() []event.Event {
	history := s.history.History()

	events := make([]event.Event, len(history))

	for i, h := range history {
		if e, ok := h.(event.Event); ok {
			events[i] = e
		}
	}

	return events

}
