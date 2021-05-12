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

package quota

import (
	"context"
	"errors"
	"github.com/xmidt-org/ears/pkg/ratelimit"
	"github.com/xmidt-org/ears/pkg/tenant"
	"time"
)

//Quota limiter implements an additional Wait function
type QuotaLimiter struct {
	tid             tenant.Id
	internalLimiter ratelimit.RateLimiter
	wakeup          chan bool
}

func NewQuotaLimiter(tid tenant.Id, redisAddr string, initialRqs int, tenantRqs int) *QuotaLimiter {
	backendLimiter := ratelimit.NewRedisRateLimiter(tid, redisAddr, tenantRqs)
	limiter := ratelimit.NewAdaptiveRateLimiter(backendLimiter, initialRqs, tenantRqs)

	return &QuotaLimiter{
		tid:             tid,
		internalLimiter: limiter,
		wakeup:          make(chan bool),
	}
}

func (r *QuotaLimiter) Wait(ctx context.Context) error {
	for {
		err := r.Take(ctx, 1)
		if err != nil {
			return nil
		}
		sleepTO := time.Second * 5
		var limitReached *ratelimit.LimitReached
		if errors.As(err, &limitReached) {
			//TODO figure out what's the optimal way of waiting
			sleepTO = time.Millisecond * 100
		}
		select {
		case <-ctx.Done():
			return nil
		case <-r.wakeup:
			//keep looping
		case <-time.After(sleepTO):
			//keep looping
		}
	}
	return nil
}

func (r *QuotaLimiter) Take(ctx context.Context, unit int) error {
	return r.internalLimiter.Take(ctx, unit)
}

func (r *QuotaLimiter) Limit() int {
	return r.internalLimiter.Limit()
}

func (r *QuotaLimiter) SetLimit(newLimit int) error {
	if r.Limit() == newLimit {
		//limit is not changed
		return nil
	}

	err := r.SetLimit(newLimit)
	if err != nil {
		return err
	}

	//none-blocking. One wakeup signal is suffice
	select {
	case r.wakeup <- true:
	default:
	}
	return nil
}
