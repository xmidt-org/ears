// SPDX-FileCopyrightText: 2025 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package filter_test

import (
	"context"
	"testing"

	"github.com/xmidt-org/ears/pkg/event"
	"github.com/xmidt-org/ears/pkg/filter"

	. "github.com/onsi/gomega"
)

func TestAdd(t *testing.T) {
	var c filter.Chain

	a := NewWithT(t)

	for i := 0; i < 5; i++ {
		err := c.Add(&filter.FiltererMock{})
		a.Expect(err).To(BeNil())

		fs := c.Filterers()
		a.Expect(len(fs)).To(Equal(i + 1))

	}

}

func TestChaining(t *testing.T) {
	testCases := []struct {
		name       string
		chain      []filter.Filterer
		multiplier int
	}{
		{
			name:       "pass",
			chain:      []filter.Filterer{newPassFilterer()},
			multiplier: 1,
		},
		{
			name:       "block",
			chain:      []filter.Filterer{newBlockFilterer()},
			multiplier: 0,
		},
		{
			name:       "double",
			chain:      []filter.Filterer{newDoubleFilterer()},
			multiplier: 2,
		},
		{
			name: "pass-pass",
			chain: []filter.Filterer{
				newPassFilterer(),
				newPassFilterer(),
			},
			multiplier: 1,
		},
		{
			name: "pass-block",
			chain: []filter.Filterer{
				newPassFilterer(),
				newBlockFilterer(),
			},
			multiplier: 0,
		},
		{
			name: "block-pass",
			chain: []filter.Filterer{
				newBlockFilterer(),
				newPassFilterer(),
			},
			multiplier: 0,
		},
		{
			name: "pass-double",
			chain: []filter.Filterer{
				newPassFilterer(),
				newDoubleFilterer(),
			},
			multiplier: 2,
		},
		{
			name: "double-pass",
			chain: []filter.Filterer{
				newDoubleFilterer(),
				newPassFilterer(),
			},
			multiplier: 2,
		},
		{
			name: "double-double",
			chain: []filter.Filterer{
				newDoubleFilterer(),
				newDoubleFilterer(),
			},
			multiplier: 4,
		},
	}

	ctx := context.Background()

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var c filter.Chain
			a := NewWithT(t)
			for _, f := range tc.chain {
				c.Add(f)
			}

			for _, payloads := range [][]string{
				{"single"},
				{"multi-1", "multi-2"},
			} {
				result := []event.Event{}
				for _, p := range payloads {
					evt, err := event.New(ctx, p, event.FailOnNack(t))
					a.Expect(err).To(BeNil())

					r := c.Filter(evt)

					if len(r) > 0 {
						result = append(result, r...)
					}
				}

				a.Expect(len(result)).To(Equal(len(payloads) * tc.multiplier))
			}
		})
	}

}

func newBlockFilterer() filter.Filterer {
	return &filter.FiltererMock{
		FilterFunc: func(e event.Event) []event.Event {
			e.Ack()
			return []event.Event{}
		},
		NameFunc: func() string {
			return "mockBlock"
		},
	}
}

func newPassFilterer() filter.Filterer {
	return &filter.FiltererMock{
		FilterFunc: func(e event.Event) []event.Event {
			return []event.Event{e}
		},
		NameFunc: func() string {
			return "mockPass"
		},
	}
}

func newDoubleFilterer() filter.Filterer {
	return &filter.FiltererMock{
		FilterFunc: func(e event.Event) []event.Event {
			e1, err := e.Clone(e.Context())
			if err != nil {
				e.Nack(err)
				return nil
			}
			e1.DeepCopy()
			e2, err := e.Clone(e.Context())
			if err != nil {
				e.Nack(err)
				return nil
			}
			e2.DeepCopy()
			e.Ack()
			return []event.Event{e1, e2}
		},
		NameFunc: func() string {
			return "mockDouble"
		},
	}
}
