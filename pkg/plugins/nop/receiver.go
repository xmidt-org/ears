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

package nop

import (
	"context"
	"fmt"
	"github.com/goccy/go-yaml"
	"github.com/rs/zerolog"
	"github.com/xmidt-org/ears/pkg/event"
	pkgplugin "github.com/xmidt-org/ears/pkg/plugin"
	"github.com/xmidt-org/ears/pkg/receiver"
	"github.com/xmidt-org/ears/pkg/secret"
	"github.com/xmidt-org/ears/pkg/tenant"
	"os"
	"time"
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
	cfg = cfg.WithDefaults()
	err = cfg.Validate()
	if err != nil {
		return nil, err
	}
	logger := zerolog.New(os.Stdout).Level(zerolog.DebugLevel)
	//zerolog.LevelFieldName = "log.level"
	r := &Receiver{
		config:     cfg,
		name:       name,
		plugin:     plugin,
		tid:        tid,
		logger:     logger,
		stopped:    true,
		currentSec: time.Now().Unix(),
	}
	return r, nil
}

func (r *Receiver) logSuccess() {
	r.Lock()
	r.successCounter++
	if time.Now().Unix() != r.currentSec {
		r.successVelocityCounter = r.currentSuccessVelocityCounter
		r.currentSuccessVelocityCounter = 0
		r.currentSec = time.Now().Unix()
	}
	r.currentSuccessVelocityCounter++
	r.Unlock()
}

func (r *Receiver) logError() {
	r.Lock()
	r.errorCounter++
	if time.Now().Unix() != r.currentSec {
		r.errorVelocityCounter = r.currentErrorVelocityCounter
		r.currentErrorVelocityCounter = 0
		r.currentSec = time.Now().Unix()
	}
	r.currentErrorVelocityCounter++
	r.Unlock()
}

func (r *Receiver) Receive(next receiver.NextFn) error {
	if r == nil {
		return &pkgplugin.Error{
			Err: fmt.Errorf("Receive called on <nil> pointer"),
		}
	}
	if next == nil {
		return &receiver.InvalidConfigError{
			Err: fmt.Errorf("next cannot be nil"),
		}
	}
	r.Lock()
	r.done = make(chan struct{})
	r.stopped = false
	r.next = next
	r.Unlock()
	<-r.done
	return nil
}

func (r *Receiver) StopReceiving(ctx context.Context) error {
	r.Lock()
	defer r.Unlock()
	if !r.stopped && r.done != nil {
		close(r.done)
		r.stopped = true
	}
	return nil
}

func (r *Receiver) Count() int {
	return 0
}

func (r *Receiver) Trigger(e event.Event) {
	// Ensure that `next` can be slow and locking here will not
	// prevent other requests from executing.
	r.Lock()
	next := r.next
	r.Unlock()
	// this will be called by RouteEvent() for webhook routes
	r.logSuccess()
	if next != nil {
		next(e)
	}
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

func (r *Receiver) EventSuccessCount() int {
	r.Lock()
	defer r.Unlock()
	return r.successCounter
}

func (r *Receiver) EventSuccessVelocity() int {
	r.Lock()
	defer r.Unlock()
	return r.successVelocityCounter
}

func (r *Receiver) EventErrorCount() int {
	r.Lock()
	defer r.Unlock()
	return r.errorCounter
}

func (r *Receiver) EventErrorVelocity() int {
	r.Lock()
	defer r.Unlock()
	return r.errorVelocityCounter
}

func (r *Receiver) EventTs() int64 {
	r.Lock()
	defer r.Unlock()
	return r.currentSec
}
