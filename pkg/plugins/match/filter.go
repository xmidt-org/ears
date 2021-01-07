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

	"github.com/goccy/go-yaml"
	"github.com/xmidt-org/ears/pkg/event"
	"github.com/xmidt-org/ears/pkg/filter"
	"github.com/xmidt-org/ears/pkg/filters"
	pkgplugin "github.com/xmidt-org/ears/pkg/plugin"
)

func NewFilter(config interface{}) (filter.Filterer, error) {

	var cfg Config
	var err error

	switch c := config.(type) {
	case string:
		err = yaml.Unmarshal([]byte(c), &cfg)

	case []byte:
		err = yaml.Unmarshal(c, &cfg)

	case Config:
		cfg = c
	case *Config:
		cfg = *c

	default:
		err = fmt.Errorf("invalid type %v", c)
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

	var matcher filters.Matcher

	switch cfg.Matcher {
	case MatcherRegex:
		matcher, err = filters.NewMatcherRegex(*cfg.Pattern)
		if err != nil {
			return nil, &pkgplugin.InvalidConfigError{
				Err: err,
			}
		}
	default:
		return nil, &pkgplugin.InvalidConfigError{
			Err: fmt.Errorf("unsupported matcher type: %s", cfg.Matcher.String()),
		}
	}

	f := &Filter{
		config: cfg,
		filter: filters.MatchFilter{Matcher: matcher},
	}

	switch cfg.Mode {
	case ModeAllow:
		f.filter.Mode = filters.MatchModeAllow
	case ModeDeny:
		f.filter.Mode = filters.MatchModeDeny
	}

	return f, nil

}

func (f *Filter) Filter(ctx context.Context, e event.Event) ([]event.Event, error) {
	if f == nil {
		return nil, &pkgplugin.InvalidConfigError{
			Err: fmt.Errorf("filter cannot be nil"),
		}
	}

	return f.filter.Filter(ctx, e)
}

func (f *Filter) Config() string {
	if f == nil {
		return `error: "nil filter"`
	}

	d, err := yaml.Marshal(f.config)
	if err != nil {
		return `error: "could not marshal configs"`
	}

	return string(d)
}
