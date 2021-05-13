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
	"golang.org/x/time/rate"
	"sync"
	"time"
)

var globalLimiters = make(map[string]*rate.Limiter)
var lock = &sync.Mutex{}

//InMemoryBackendLimiter is only for testing/unit test purpose
type InMemoryBackendLimiter struct {
	tid           tenant.Id
	tenantLimiter *rate.Limiter
}

func NewInMemoryBackendLimiter(tid tenant.Id, rqs int) *InMemoryBackendLimiter {
	lock.Lock()
	defer lock.Unlock()

	limiter, ok := globalLimiters[tid.Key()]
	if !ok {
		limiter = rate.NewLimiter(rate.Limit(rqs), 1)
		globalLimiters[tid.Key()] = limiter
	}
	return &InMemoryBackendLimiter{
		tid,
		limiter,
	}
}

func (r *InMemoryBackendLimiter) Take(ctx context.Context, unit int) error {
	allowed := r.tenantLimiter.AllowN(time.Now(), unit)
	if !allowed {
		return &LimitReached{}
	}
	return nil
}

func (r *InMemoryBackendLimiter) Limit() int {
	return int(r.tenantLimiter.Limit())
}

func (r *InMemoryBackendLimiter) SetLimit(newLimit int) error {
	if newLimit < 0 {
		return &InvalidUnitError{newLimit}
	}
	r.tenantLimiter.SetLimit(rate.Limit(newLimit))
	return nil
}
