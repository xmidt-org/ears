package chain

import (
	"context"
	"fmt"
	"sync"

	"github.com/xmidt-org/ears/pkg/event"
	"github.com/xmidt-org/ears/pkg/filter"
)

func (c *Chain) Add(f filter.Filterer) error {
	if f == nil {
		return &InvalidArgumentError{
			Err: fmt.Errorf("filter cannot be nil"),
		}
	}

	c.Lock()
	defer c.Unlock()
	if c.filterers == nil {
		c.filterers = make([]filter.Filterer, 1)
	}

	c.filterers = append(c.filterers, f)

	return nil
}

func (c *Chain) Filter(ctx context.Context, e event.Event) ([]event.Event, []error) {

	c.Lock()
	defer c.Unlock()

	var errs = []error{}
	var events = []event.Event{e}

	for _, f := range c.filterers {
		numEvents := len(events)

		switch numEvents {
		case 0:
			return nil, nil

		case 1:
			events, errs = f.Filter(ctx, e)
			if len(errs) > 0 {
				return nil, errs
			}

		default:
			errArrCh := make(chan []error, numEvents)
			eventArrCh := make(chan []event.Event, numEvents)
			var wg sync.WaitGroup

			wg.Add(len(events))
			for _, e := range events {
				go func(e event.Event) {
					events, errs := f.Filter(ctx, e)
					if len(events) > 0 {
						eventArrCh <- events
					}
					if len(errs) > 0 {
						errArrCh <- errs
					}
					wg.Done()
				}(e)
			}

			wg.Wait()
			close(errArrCh)
			close(eventArrCh)

			if len(errArrCh) > 0 {
				for errArr := range errArrCh {
					errs = append(errs, errArr...)
				}

				return nil, errs
			}

			events = []event.Event{}
			for evtArr := range eventArrCh {
				events = append(events, evtArr...)
			}

		}
	}

	return events, nil

}
