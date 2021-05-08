package ratelimit

import (
	"context"
	"errors"
	"github.com/rs/zerolog/log"
	"golang.org/x/time/rate"
	"math"
	"sync"
	"time"
)

type AdaptiveRateLimiter struct {
	backend    RateLimiter
	initialRqs int
	totalRqs   int
	currentRqs int
	limiter    *rate.Limiter
	lock       *sync.Mutex
	lastTune   time.Time
	takeCount  int
	limitCount int
	workerName string
}

func NewAdaptiveRateLimiter(workerName string, backendLimiter RateLimiter, initialRqs int, totalRqs int) *AdaptiveRateLimiter {
	return &AdaptiveRateLimiter{
		backend:    backendLimiter,
		initialRqs: initialRqs,
		totalRqs:   totalRqs,
		currentRqs: -1,
		limiter:    nil,
		lock:       &sync.Mutex{},
		lastTune:   time.Now(),
		takeCount:  0,
		limitCount: 0,
		workerName: workerName, //debug only
	}
}

func (r *AdaptiveRateLimiter) SetLimit(newLimit int) error {
	if newLimit < 0 {
		return &InvalidUnitError{newLimit}
	}

	r.lock.Lock()
	defer r.lock.Unlock()

	r.totalRqs = newLimit
	return nil
}

func (r *AdaptiveRateLimiter) Limit() int {
	return r.totalRqs
}

func (r *AdaptiveRateLimiter) Take(ctx context.Context, unit int) error {
	r.lock.Lock()
	defer r.lock.Unlock()

	if r.totalRqs == 0 {
		//no limit
		return nil
	}

	var err error
	if r.limiter == nil {
		err = r.initLimiter(ctx)
	} else {
		err = r.tuneRqs(ctx)
	}
	if err != nil {
		return err
	}

	allowed := r.limiter.AllowN(time.Now(), unit)
	if !allowed {
		r.limitCount++
		return &LimitReached{}
	}
	r.takeCount++
	return nil
}

func (r *AdaptiveRateLimiter) initLimiter(ctx context.Context) error {

	err := r.backend.Take(ctx, r.initialRqs)
	if err != nil {
		r.initialRqs = r.initialRqs / 2
		if r.initialRqs == 0 {
			//rqs should never be below 1 rqs (for now)
			r.initialRqs = 1
		}
		return &BackendError{err}
	}
	r.limiter = rate.NewLimiter(rate.Limit(r.initialRqs), 1)
	r.currentRqs = r.initialRqs
	r.lastTune = time.Now()
	return nil
}

//Tune RQS check the rate limit history and ask backend ratelimiter for a new quota if necessary
func (r *AdaptiveRateLimiter) tuneRqs(ctx context.Context) error {
	if r.lastTune.Add(time.Second).After(time.Now()) {
		//No need to tune yet
		return nil
	}

	//fmt.Printf("%p, tune last=%s, now=%s\n", r.backend, ts(r.lastTune), ts(time.Now()))

	//fmt.Printf("(%s) rqs: %d, takeCount: %d, limitCount: %d\n", r.workerName, r.currentRqs, r.takeCount, r.limitCount)

	newRqs := r.currentRqs
	if r.limitCount > 0 {
		if newRqs*2 < r.totalRqs {
			//looks like there are room to grow, lets ask for more quota
			newRqs = newRqs * 2
		}
	} else if r.takeCount < newRqs {
		takeCount := float32(r.takeCount)
		newRqs = int(takeCount * 1.2)

		diffPercentage := math.Abs(float64(newRqs-r.currentRqs)) / float64(newRqs)
		if diffPercentage < 0.05 {
			//not much difference between the new rqs vs older one. Don't bother update
			newRqs = r.currentRqs
		}
	}

	//try to see if we can actually take newRqs from backend
	target := newRqs
	floor := r.currentRqs

	for {
		err := r.backend.Take(ctx, target)
		if err == nil {
			//we are good
			newRqs = target
			break
		}
		var limitReached *LimitReached
		if !errors.As(err, &limitReached) {
			return &BackendError{err}
		}
		if target <= floor {
			target = floor
			floor = 0
		}
		//fmt.Printf("(%s) target: %d, floor: %d\n", r.workerName, target, floor)
		target = (target + floor) / 2
		if target == 0 {
			return &BackendError{err}
		}
	}

	r.lastTune = time.Now()
	r.takeCount = 0
	r.limitCount = 0

	if newRqs == r.currentRqs {
		return nil
	}

	log.Ctx(ctx).Info().Str("op", "tuneRqs").Int("NewRqs", target).Int("OldRqs", r.currentRqs).
		Msg("Updating new ratelimit")

	r.limiter.SetLimit(rate.Limit(newRqs))
	r.currentRqs = newRqs
	return nil
}
