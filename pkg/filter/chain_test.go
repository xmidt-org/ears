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
					evt, err := event.NewEvent(p)
					a.Expect(err).To(BeNil())

					r, err := c.Filter(ctx, evt)
					a.Expect(err).To(BeNil())

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
		FilterFunc: func(ctx context.Context, e event.Event) ([]event.Event, error) {
			return []event.Event{}, nil
		},
	}
}

func newPassFilterer() filter.Filterer {
	return &filter.FiltererMock{
		FilterFunc: func(ctx context.Context, e event.Event) ([]event.Event, error) {
			return []event.Event{e}, nil
		},
	}
}

func newDoubleFilterer() filter.Filterer {
	return &filter.FiltererMock{
		FilterFunc: func(ctx context.Context, e event.Event) ([]event.Event, error) {
			return []event.Event{e, e}, nil
		},
	}
}
