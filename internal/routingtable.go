package internal

import (
	"context"
	"crypto/md5"
	"fmt"
)

func (rte *RoutingTableEntry) Hash(ctx context.Context) string {
	str := rte.Source.Hash(ctx) + rte.Destination.Hash(ctx)
	if rte.FilterChain != nil {
		for _, filter := range rte.FilterChain {
			str += filter.Hash(ctx)
		}
	}
	hash := fmt.Sprintf("%x", md5.Sum([]byte(str)))
	return hash
}

func (rte *RoutingTableEntry) Validate(ctx context.Context) error {
	return nil
}

func (rte *RoutingTableEntry) Initialize(ctx context.Context) error {
	//
	// initialize input plugin
	//
	var err error
	rte.Source, err = NewInputPlugin(ctx, rte)
	if err != nil {
		return err
	}
	//
	// initialize filter chain
	//
	// for now input an output plugin are not connected via channel but rather via function call
	// therefore the first filter in the filer chain has a nil input chain and the last filter in the filter chain has a nil output chain
	// we will likely change this to channel in the future
	var eventChannel chan *Event
	if rte.FilterChain != nil {
		for idx, fp := range rte.FilterChain {
			var err error
			fp.Filterer, err = NewFilterer(ctx, fp)
			if err != nil {
				return err
			}
			fp.InputChannel = eventChannel
			if idx < len(rte.FilterChain)-1 {
				fp.OutputChannel = make(chan *Event)
				eventChannel = fp.OutputChannel
			}
		}
	}
	//
	// initialize output plugin
	//
	rte.Destination, err = NewOutputPlugin(ctx, rte)
	if err != nil {
		return err
	}
	return nil
}

func (fp *FilterPlugin) DoSync(ctx context.Context, event *Event) error {
	filteredEvents, err := fp.Filterer.Filter(ctx, event)
	if err != nil {
		return err
	}
	for _, e := range filteredEvents {
		if fp.OutputChannel != nil {
			fp.OutputChannel <- e
		} else {
			//TODO pass event to output plugin
		}
	}
	return nil
}

func (fp *FilterPlugin) DoAsync(ctx context.Context) {
	go func() {
		if fp.InputChannel == nil {
			return
		}
		for {
			inputEvent := <-fp.InputChannel
			filteredEvents, err := fp.Filterer.Filter(ctx, inputEvent)
			if err != nil {
				return
			}
			for _, e := range filteredEvents {
				if fp.OutputChannel != nil {
					fp.OutputChannel <- e
				} else {
					//TODO pass event to output plugin
				}
			}
		}
	}()
}
