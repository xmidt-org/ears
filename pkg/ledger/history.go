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

package ledger

import (
	"container/ring"
)

func NewLimitedRecorder(size int) *History {
	h := History{
		size: size,
	}

	if size > 0 {
		h.ring = ring.New(size)
	}

	return &h
}

func (h *History) Add(e interface{}) error {
	h.Lock()
	defer h.Unlock()

	h.count++

	if h.ring == nil {
		return &NotInitializedError{}
	}

	h.ring.Value = e
	h.ring = h.ring.Next()

	return nil
}

func (h *History) Count() int {
	return h.count
}

func (h *History) Capacity() int {
	return h.size
}

func (h *History) Records() ([]interface{}, error) {
	h.Lock()
	defer h.Unlock()

	if h.ring == nil {
		return nil, &NotInitializedError{}
	}

	move := h.count * -1
	iterations := h.count
	if h.count > h.size {
		move = 0
		iterations = h.size
	}

	r := h.ring.Move(move)

	events := make([]interface{}, iterations)

	for i := 0; i < iterations; i++ {
		events[i] = r.Value
		r = r.Next()
	}

	return events, nil

}
