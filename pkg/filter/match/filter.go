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
	"context"
	"fmt"

	"github.com/xmidt-org/ears/pkg/event"
	"github.com/xmidt-org/ears/pkg/filter"
	pkgregex "github.com/xmidt-org/ears/pkg/filter/match/regex"
)

// Ensure supporting matchers implement Matcher interface
var _ Matcher = (*pkgregex.Matcher)(nil)

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

	var matcher Matcher

	switch cfg.Matcher {
	case MatcherRegex:
		matcher, err = pkgregex.NewMatcher(*cfg.Pattern)
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
		matcher: matcher,
	}

	return f, nil
}

func (f *Filter) Filter(ctx context.Context, evt event.Event) ([]event.Event, error) {
	if f == nil {
		return nil, &filter.InvalidConfigError{
			Err: fmt.Errorf("<nil> pointer filter"),
		}
	}

	// passes if event matches
	events := []event.Event{}
	pass := f.matcher.Match(ctx, evt)

	if f.config.Mode == ModeDeny {
		pass = !pass
	}

	if pass {
		events = []event.Event{evt}
	}

	return events, nil
}

func (f *Filter) Config() Config {
	if f == nil {
		return Config{}
	}

	return f.config
}
