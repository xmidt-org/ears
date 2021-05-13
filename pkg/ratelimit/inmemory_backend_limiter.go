// Copyright 2021 Comcast Cable Communications Management, LLC
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

package ratelimit

import (
	"context"
	"github.com/xmidt-org/ears/pkg/tenant"
	"sync"
	"time"
)

var globalLimiters = make(map[string]*InMemoryBackendLimiter)
var lock = &sync.Mutex{}

//InMemoryBackendLimiter is only for testing/unit test purpose
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
