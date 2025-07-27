// SPDX-FileCopyrightText: 2025 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package debug

import (
	"container/ring"
)

func newHistory(size int) *history {
	h := history{
		size: size,
	}

	if size > 0 {
		h.ring = ring.New(size)
	}

	return &h
}

func (h *history) Add(e interface{}) {
	h.Lock()
	defer h.Unlock()

	h.count++

	if h.ring == nil {
		return
	}

	h.ring.Value = e
	h.ring = h.ring.Next()
}

func (h *history) Count() int {
	return h.count
}

func (h *history) Size() int {
	return h.size
}

func (h *history) History() []interface{} {
	h.Lock()
	defer h.Unlock()

	if h.ring == nil {
		return []interface{}{}
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

	return events

}
