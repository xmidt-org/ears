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

package dedup

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	lru "github.com/hashicorp/golang-lru"
	"github.com/xmidt-org/ears/pkg/event"
	"github.com/xmidt-org/ears/pkg/filter"
)

func NewFilter(config interface{}) (*Filter, error) {
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
	}
	f.lruCache, err = lru.New(*cfg.CacheSize)
	return f, nil
}

func (f *Filter) Filter(evt event.Event) []event.Event {
	if f == nil {
		evt.Nack(&filter.InvalidConfigError{
			Err: fmt.Errorf("<nil> pointer filter"),
		})
		return nil
	}
	obj, _, _ := evt.GetPathValue(f.config.DedupPath)
	buf, err := json.Marshal(obj)
	if err != nil {
		evt.Nack(err)
		return []event.Event{}
	}
	evtHash := fmt.Sprintf("%x", md5.Sum(buf))
	_, ok := f.lruCache.Get(evtHash)
	if !ok {
		f.lruCache.Add(evtHash, struct{}{})
		return []event.Event{evt}
	} else {
		evt.Ack()
		return []event.Event{}
	}
}

func (f *Filter) Config() Config {
	if f == nil {
		return Config{}
	}
	return f.config
}
