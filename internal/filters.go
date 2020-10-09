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
		Pattern interface{} // match pattern
		Matcher Matcher
	}

	FilterFilter struct {
		Pattern interface{} // filter pattern
		Matcher Matcher
	}

	TransformFilter struct {
		Transformation interface{} // transformation instructions for event payload
	}

	// a TTLFilter filters an event when it is too old
	TTLFilter struct {
		TTL    int    // ttl in seconds
		TsPath string // path to timestamp information in payload
	}

	// a SplitFilter splits and event into two or more events
	SplitFilter struct {
		SplitPath string // path to split array in payload
	}

	// a PassFilter always passes an incoming event
	PassFilter struct {
	}

	// a BlockFilter always blocks an incoming event
	BlockFilter struct {
	}
)

// Filter filters events not matching a pattern
func (mf *MatchFilter) Filter(ctx context.Context, event *Event) ([]*Event, error) {
	// passes if event matches
	events := make([]*Event, 0)
	if mf.Matcher.Match(ctx, event, mf.Pattern) {
		events = append(events, event)
	}
	return events, nil
}

// Filter filters events matching a pattern
func (mf *FilterFilter) Filter(ctx context.Context, event *Event) ([]*Event, error) {
	// passes if event does not match
	events := make([]*Event, 0)
	if !mf.Matcher.Match(ctx, event, mf.Pattern) {
		events = append(events, event)
	}
	return events, nil
}

// Filter transforms an event according to its transformation spec
func (mf *TransformFilter) Filter(ctx context.Context, event *Event) ([]*Event, error) {
	// identity transform
	events := make([]*Event, 0)
	//TODO: implement filter logic
	events = append(events, event)
	return events, nil
}

// Filter filters event if it hhas expired
func (mf *TTLFilter) Filter(ctx context.Context, event *Event) ([]*Event, error) {
	// never filters
	events := make([]*Event, 0)
	//TODO: implement filter logic
	events = append(events, event)
	return events, nil
}

// Filter splits an event containing an array into multiple events
func (mf *SplitFilter) Filter(ctx context.Context, event *Event) ([]*Event, error) {
	// always splits into two identical events
	events := make([]*Event, 0)
	//TODO: implement filter logic
	//TODO: clone
	events = append(events, event)
	events = append(events, event)
	return events, nil
}

// Filter lets any event pass
func (mf *PassFilter) Filter(ctx context.Context, event *Event) ([]*Event, error) {
	return []*Event{event}, nil
}

// Filter lets no event pass
func (mf *BlockFilter) Filter(ctx context.Context, event *Event) ([]*Event, error) {
	return []*Event{}, nil
}

// NewFilterer is a factory function to create appropriate filterer for given filter plugin config
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
	case FilterTypePass:
		return new(PassFilter), nil
	case FilterTypeBlock:
		return new(BlockFilter), nil
	}
	return nil, errors.New("unknown filter type " + fp.Type)
}

func (fp *FilterPlugin) Close() {
	fp.done <- true
}

// DoSync synchronoulsy accepts an event into a filter plugin belonging to a filter chain
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

// DoAsync kicks of a go routine listening indefinitly for a events to arrive on its in channel
func (fp *FilterPlugin) DoAsync(ctx context.Context) {
	go func() {
		if fp.inputChannel == nil {
			return
		}
		for {
			var inputEvent *Event
			select {
			case inputEvent = <-fp.inputChannel:
			case <-fp.done:
				return
			}
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
