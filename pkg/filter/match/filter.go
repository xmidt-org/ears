// SPDX-FileCopyrightText: 2025 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

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
		config:  *cfg,
		name:    name,
		plugin:  plugin,
		tid:     tid,
		matcher: matcher,
	}
	f.MetricFilter = filter.NewMetricFilter(tableSyncer, f.Hash)
	return f, nil
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
		f.LogSuccess()
		events = []event.Event{evt}
	} else {
		f.LogFilter()
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
