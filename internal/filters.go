/**
 *  Copyright (c) 2020  Comcast Cable Communications Management, LLC
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package internal

import (
	"context"
	"errors"

	"github.com/rs/zerolog/log"
)

type (
	MatchFilter struct {
		Pattern interface{}
		Matcher Matcher
	}

	FilterFilter struct {
		Pattern interface{}
		Matcher Matcher
	}

	TransformFilter struct {
		Transformation interface{}
	}

	TTLFilter struct {
		TTL    int
		TsPath string
	}

	SplitFilter struct {
		SplitPath string
	}
)

func (mf *MatchFilter) Filter(ctx context.Context, event *Event) ([]*Event, error) {
	// passes if event matches
	events := make([]*Event, 0)
	if mf.Matcher.Match(ctx, event, mf.Pattern) {
		events = append(events, event)
	}
	return events, nil
}

func (mf *FilterFilter) Filter(ctx context.Context, event *Event) ([]*Event, error) {
	// passes if event does not match
	events := make([]*Event, 0)
	if !mf.Matcher.Match(ctx, event, mf.Pattern) {
		events = append(events, event)
	}
	return events, nil
}

func (mf *TransformFilter) Filter(ctx context.Context, event *Event) ([]*Event, error) {
	// identity transform
	events := make([]*Event, 0)
	//TODO: implement filter logic
	events = append(events, event)
	return events, nil
}

func (mf *TTLFilter) Filter(ctx context.Context, event *Event) ([]*Event, error) {
	// never filters
	events := make([]*Event, 0)
	//TODO: implement filter logic
	events = append(events, event)
	return events, nil
}

func (mf *SplitFilter) Filter(ctx context.Context, event *Event) ([]*Event, error) {
	// always splits into two identical events
	events := make([]*Event, 0)
	//TODO: implement filter logic
	events = append(events, event)
	events = append(events, event)
	return events, nil
}

// factory function to create appropriate filterer for given filter plugin config
func NewFilterer(ctx context.Context, fp *FilterPlugin) (Filterer, error) {
	if fp == nil {
		return nil, errors.New("missing filter plugin config")
	}
	switch fp.Type {
	case FilterTypeFilter:
		flt := new(FilterFilter)
		flt.Matcher = NewDefaultPatternMatcher()
		// parse config params
		if fp.Params != nil {
			if value, ok := fp.Params["pattern"]; ok {
				flt.Pattern = value
			}
		}
		return flt, nil
	case FilterTypeMatcher:
		flt := new(MatchFilter)
		flt.Matcher = NewDefaultPatternMatcher()
		// parse config params
		if fp.Params != nil {
			if value, ok := fp.Params["pattern"]; ok {
				flt.Pattern = value
			}
		}
		return flt, nil
	case FilterTypeTransformer:
		//TODO: pass in config params here
		return new(TransformFilter), nil
	case FilterTypeTTLer:
		//TODO: pass in config params here
		return new(TTLFilter), nil
	case FilterTypeSplitter:
		//TODO: pass in config params here
		return new(SplitFilter), nil
	}
	return nil, errors.New("unknown filter type " + fp.Type)
}

func (fp *FilterPlugin) DoSync(ctx context.Context, event *Event) error {
	log.Debug().Msg(fp.Type + " filter " + fp.Hash(ctx) + " passed")
	filteredEvents, err := fp.filterer.Filter(ctx, event)
	if err != nil {
		return err
	}
	for _, e := range filteredEvents {
		if fp.outputChannel != nil {
			fp.outputChannel <- e
		} else {
			fp.routingTableEntry.Destination.DoSync(ctx, e)
		}
	}
	return nil
}

func (fp *FilterPlugin) DoAsync(ctx context.Context) {
	go func() {
		if fp.inputChannel == nil {
			return
		}
		for {
			inputEvent := <-fp.inputChannel
			log.Debug().Msg(fp.Type + " filter " + fp.Hash(ctx) + " passed")
			filteredEvents, err := fp.filterer.Filter(ctx, inputEvent)
			if err != nil {
				return
			}
			for _, e := range filteredEvents {
				if fp.outputChannel != nil {
					fp.outputChannel <- e
				} else {
					fp.routingTableEntry.Destination.DoSync(ctx, e)
				}
			}
		}
	}()
}
