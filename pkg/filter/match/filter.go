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

package match

import (
	"encoding/json"
	"fmt"
	"github.com/rs/zerolog/log"
	"github.com/xmidt-org/ears/internal/pkg/syncer"
	"github.com/xmidt-org/ears/pkg/event"
	"github.com/xmidt-org/ears/pkg/filter"
	"github.com/xmidt-org/ears/pkg/filter/match/comparison"
	"github.com/xmidt-org/ears/pkg/filter/match/pattern"
	"github.com/xmidt-org/ears/pkg/filter/match/patternregex"
	"github.com/xmidt-org/ears/pkg/filter/match/regex"
	"github.com/xmidt-org/ears/pkg/hasher"
	"github.com/xmidt-org/ears/pkg/secret"
	"github.com/xmidt-org/ears/pkg/tenant"
	"time"
)

// Ensure supporting matchers implement Matcher interface
var _ Matcher = (*regex.Matcher)(nil)

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
	var matcher Matcher
	switch cfg.Matcher {
	case MatcherRegex:
		matcher, err = regex.NewMatcher(cfg.Pattern, cfg.Path)
		if err != nil {
			return nil, &filter.InvalidConfigError{
				Err: err,
			}
		}
	case MatcherPattern:
		matcher, err = pattern.NewMatcher(cfg.Pattern, cfg.Patterns, cfg.PatternsLogic, *cfg.ExactArrayMatch, cfg.Path)
		if err != nil {
			return nil, &filter.InvalidConfigError{
				Err: err,
			}
		}
	case MatcherPatternRegex:
		matcher, err = patternregex.NewMatcher(cfg.Pattern, cfg.Patterns, cfg.PatternsLogic, *cfg.ExactArrayMatch, cfg.Path)
		if err != nil {
			return nil, &filter.InvalidConfigError{
				Err: err,
			}
		}
	case MatcherComparison:
		matcher, err = comparison.NewMatcher(cfg.ComparisonTree, cfg.Comparison, cfg.PatternsLogic)
		if err != nil {
			return nil, &filter.InvalidConfigError{
				Err: err,
			}
		}
	default:
		return nil, &filter.InvalidConfigError{
			Err: fmt.Errorf("unsupported matcher type: %s", cfg.Matcher.String()),
		}
	}
	f := &Filter{
		config:      *cfg,
		name:        name,
		plugin:      plugin,
		tid:         tid,
		matcher:     matcher,
		currentSec:  time.Now().Unix(),
		tableSyncer: tableSyncer,
	}
	return f, nil
}

func (f *Filter) logSuccess() {
	f.Lock()
	f.successCounter++
	if time.Now().Unix() != f.currentSec {
		f.successVelocityCounter = f.currentSuccessVelocityCounter
		f.currentSuccessVelocityCounter = 0
		f.currentSec = time.Now().Unix()
	}
	f.currentSuccessVelocityCounter++
	f.Unlock()
}

func (f *Filter) logError() {
	f.Lock()
	f.errorCounter++
	if time.Now().Unix() != f.currentSec {
		f.errorVelocityCounter = f.currentErrorVelocityCounter
		f.currentErrorVelocityCounter = 0
		f.currentSec = time.Now().Unix()
	}
	f.currentErrorVelocityCounter++
	f.Unlock()
}

func (f *Filter) logFilter() {
	f.Lock()
	f.filterCounter++
	if time.Now().Unix() != f.currentSec {
		f.filterVelocityCounter = f.currentFilterVelocityCounter
		f.currentFilterVelocityCounter = 0
		f.currentSec = time.Now().Unix()
	}
	f.currentFilterVelocityCounter++
	f.Unlock()
}

func (f *Filter) Filter(evt event.Event) []event.Event {
	if f == nil {
		evt.Nack(&filter.InvalidConfigError{
			Err: fmt.Errorf("<nil> pointer filter"),
		})
		return nil
	}
	// passes if event matches
	events := []event.Event{}
	pass := f.matcher.Match(evt)
	if f.config.Mode == ModeDeny {
		pass = !pass
	}
	if pass {
		f.logSuccess()
		events = []event.Event{evt}
	} else {
		f.logFilter()
		evt.Ack()
	}
	log.Ctx(evt.Context()).Debug().Str("op", "filter").Str("filterType", "match").Str("name", f.Name()).Int("eventCount", len(events)).Msg("match")
	return events
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

func (f *Filter) getLocalMetric() *syncer.EarsMetric {
	f.Lock()
	defer f.Unlock()
	metrics := &syncer.EarsMetric{
		f.successCounter,
		f.errorCounter,
		f.filterCounter,
		f.successVelocityCounter,
		f.errorVelocityCounter,
		f.filterVelocityCounter,
		f.currentSec,
		0,
	}
	return metrics
}

func (f *Filter) EventSuccessCount() int {
	hash := f.Hash()
	f.tableSyncer.WriteMetrics(hash, f.getLocalMetric())
	return f.tableSyncer.ReadMetrics(hash).SuccessCount
}

func (f *Filter) EventSuccessVelocity() int {
	hash := f.Hash()
	f.tableSyncer.WriteMetrics(hash, f.getLocalMetric())
	return f.tableSyncer.ReadMetrics(hash).SuccessVelocity
}

func (f *Filter) EventFilterCount() int {
	hash := f.Hash()
	f.tableSyncer.WriteMetrics(hash, f.getLocalMetric())
	return f.tableSyncer.ReadMetrics(hash).FilterCount
}

func (f *Filter) EventFilterVelocity() int {
	hash := f.Hash()
	f.tableSyncer.WriteMetrics(hash, f.getLocalMetric())
	return f.tableSyncer.ReadMetrics(hash).FilterVelocity
}

func (f *Filter) EventErrorCount() int {
	hash := f.Hash()
	f.tableSyncer.WriteMetrics(hash, f.getLocalMetric())
	return f.tableSyncer.ReadMetrics(hash).ErrorCount
}

func (f *Filter) EventErrorVelocity() int {
	hash := f.Hash()
	f.tableSyncer.WriteMetrics(hash, f.getLocalMetric())
	return f.tableSyncer.ReadMetrics(hash).ErrorVelocity
}

func (f *Filter) EventTs() int64 {
	hash := f.Hash()
	f.tableSyncer.WriteMetrics(hash, f.getLocalMetric())
	return f.tableSyncer.ReadMetrics(hash).LastEventTs
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
