// SPDX-FileCopyrightText: 2025 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package ratelimit

import (
	"context"
	"sync"
	"time"

	"github.com/xmidt-org/ears/pkg/tenant"
)

var globalLimiters = make(map[string]*InMemoryBackendLimiter)
var lock = &sync.Mutex{}

// InMemoryBackendLimiter is only for testing/unit test purpose
type InMemoryBackendLimiter struct {
	sync.Mutex

	rqs  int       // request per second
	last time.Time // last time we were polled/asked

	allowance float64
}

func NewInMemoryBackendLimiter(tid tenant.Id, rqs int) *InMemoryBackendLimiter {
	lock.Lock()
	defer lock.Unlock()

	limiter, ok := globalLimiters[tid.Key()]
	if ok {
		return limiter
	}

	limiter = &InMemoryBackendLimiter{rqs: rqs, last: time.Now()}
	limiter.allowance = float64(rqs)
	globalLimiters[tid.Key()] = limiter

	return limiter
}

func (r *InMemoryBackendLimiter) Take(ctx context.Context, unit int) error {
	if r.rqs == 0 {
		return &LimitReached{}
	}

	r.Lock()
	defer r.Unlock()

	rate := float64(r.rqs)
	now := time.Now()
	elapsed := now.Sub(r.last)
	r.last = now
	r.allowance += elapsed.Seconds() * rate

	if r.allowance > rate {
		r.allowance = rate
	}

	if r.allowance < float64(unit) {
		return &LimitReached{}
	}
	r.allowance -= float64(unit)
	return nil
}

func (r *InMemoryBackendLimiter) Limit() int {
	return r.rqs
}

func (r *InMemoryBackendLimiter) SetLimit(newLimit int) error {
	if newLimit < 0 {
		return &InvalidUnitError{newLimit}
	}
	r.rqs = newLimit
	return nil
}
