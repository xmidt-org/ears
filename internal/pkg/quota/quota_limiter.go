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
	"github.com/rs/zerolog/log"
	"github.com/xmidt-org/ears/pkg/ratelimit"
	"github.com/xmidt-org/ears/pkg/ratelimit/redis"
	"github.com/xmidt-org/ears/pkg/tenant"
	"time"
)

//Quota limiter implements an additional Wait function
type QuotaLimiter struct {
	tid             tenant.Id
	adaptiveLimiter *ratelimit.AdaptiveRateLimiter
	wakeup          chan bool
}

func NewQuotaLimiter(tid tenant.Id, backendLimiterType string, redisAddr string, initialRqs int, tenantRqs int) *QuotaLimiter {

	var backendLimiter ratelimit.RateLimiter
	if backendLimiterType == "redis" {
		backendLimiter = redis.NewRedisRateLimiter(tid, redisAddr, tenantRqs, 3)
	} else {
		backendLimiter = ratelimit.NewInMemoryBackendLimiter(tid, tenantRqs)
	}
	limiter := ratelimit.NewAdaptiveRateLimiter(backendLimiter, initialRqs, tenantRqs)

	return &QuotaLimiter{
		tid:             tid,
		adaptiveLimiter: limiter,
		wakeup:          make(chan bool),
	}
}

func (r *QuotaLimiter) Wait(ctx context.Context) error {
	for {
		err := r.Take(ctx, 1)
		if err == nil {
			return nil
		}
		sleepTO := time.Second * 5
		var limitReached *ratelimit.LimitReached
		if errors.As(err, &limitReached) {
			//TODO figure out what's the optimal way of waiting
			sleepTO = time.Millisecond * 100
		} else {
			log.Ctx(ctx).Error().Str("op", "QuotaLimiter.Wait").Str("error", err.Error()).Msg("Error taking quota")
		}
		select {
		case <-ctx.Done():
			return &ratelimit.ContextCancelled{}
		case <-r.wakeup:
			//keep looping
		case <-time.After(sleepTO):
			//keep looping
		}
	}
}

func (r *QuotaLimiter) Take(ctx context.Context, unit int) error {
	return r.adaptiveLimiter.Take(ctx, unit)
}

func (r *QuotaLimiter) Limit() int {
	return r.adaptiveLimiter.Limit()
}

func (r *QuotaLimiter) AdaptiveLimit() int {
	return r.adaptiveLimiter.AdaptiveLimit()
}

func (r *QuotaLimiter) SetLimit(newLimit int) error {
	if r.Limit() == newLimit {
		//limit is not changed
		return nil
	}

	err := r.adaptiveLimiter.SetLimit(newLimit)
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
