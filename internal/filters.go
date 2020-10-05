package internal

import (
	"context"
	"errors"
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
