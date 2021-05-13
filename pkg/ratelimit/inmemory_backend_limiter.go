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
