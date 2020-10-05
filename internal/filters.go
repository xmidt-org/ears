package internal

import (
	"context"
	"errors"
)

const (
	FilterTypeFilter      = "filter"
	FilterTypeMatcher     = "matcher"
	FilterTypeTransformer = "transformer"
	FilterTypeTTLer       = "ttler"
	FilterTypeSplitter    = "splitter"
)

type (
	MatchFilter struct {
		Pattern interface{}
	}

	FilterFilter struct {
		Pattern interface{}
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
	// always matches
	events := make([]*Event, 0)
	//TODO: implement filter logic
	events = append(events, event)
	return events, nil
}

func (mf *FilterFilter) Filter(ctx context.Context, event *Event) ([]*Event, error) {
	// never filters
	events := make([]*Event, 0)
	//TODO: implement filter logic
	events = append(events, event)
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
	//TODO: pass in config params here
	switch fp.FilterType {
	case FilterTypeFilter:
		return new(FilterFilter), nil
	case FilterTypeMatcher:
		return new(MatchFilter), nil
	case FilterTypeTransformer:
		return new(TransformFilter), nil
	case FilterTypeTTLer:
		return new(TTLFilter), nil
	case FilterTypeSplitter:
		return new(SplitFilter), nil
	}
	return nil, errors.New("unknown filter type " + fp.FilterType)
}
