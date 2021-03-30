package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/go-redis/redis/v8"
	"golang.org/x/time/rate"
	"math/rand"
	"strconv"
	"sync"
	"time"
)

type RateLimiter interface {

	//Take given units
	//Returns LimitReached if limit reached
	//Returns InvalidUnitError if unit <= 0 or > the max limit
	Take(ctx context.Context, unit int) error

	//Wait until units are available
	//Returns InvalidUnitError if unit <= 0 or > the max limit
	//Returns CancelledError if cancelled by context
	//Wait(ctx context.Context, unit int) error
}

type AdaptiveRateLimiter struct {
	backend    RateLimiter
	initialRqs int
	maxRqs     int
	currentRqs int
	limiter    *rate.Limiter
	lock       *sync.Mutex
	lastTune   time.Time
	takeCount  int
	limitCount int
	workerName string
}

func NewAdaptiveRateLimiter(workerName string, backendLimiter RateLimiter, initialRqs int, maxRqs int) *AdaptiveRateLimiter {
	return &AdaptiveRateLimiter{
		backend:    backendLimiter,
		initialRqs: initialRqs,
		maxRqs:     maxRqs,
		currentRqs: -1,
		limiter:    nil,
		lock:       &sync.Mutex{},
		lastTune:   time.Now(),
		takeCount:  0,
		limitCount: 0,
		workerName: workerName, //debug only
	}
}

func (r *AdaptiveRateLimiter) Take(ctx context.Context, unit int) error {
	r.lock.Lock()
	defer r.lock.Unlock()

	var err error
	if r.limiter == nil {
		err = r.initLimiter(ctx)
	} else {
		err = r.tuneRqs()
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
func (r *AdaptiveRateLimiter) tuneRqs() error {
	if r.lastTune.Add(time.Second).After(time.Now()) {
		//No need to tune yet
		return nil
	}

	//fmt.Printf("%p, tune last=%s, now=%s\n", r.backend, ts(r.lastTune), ts(time.Now()))

	//fmt.Printf("(%s) rqs: %d, takeCount: %d, limitCount: %d\n", r.workerName, r.currentRqs, r.takeCount, r.limitCount)

	newRqs := r.currentRqs
	if r.limitCount > 0 {
		if newRqs*2 < r.maxRqs {
			//looks like there are room to grow, lets ask for more quota
			newRqs = newRqs * 2
		}
	} else if r.takeCount < newRqs {
		//aggressively shrink
		newRqs = r.takeCount
	}
	ctx := context.Background()

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

	fmt.Printf("(%s) new rqs: %d, old rqs: %d\n", r.workerName, newRqs, r.currentRqs)

	r.limiter.SetLimit(rate.Limit(newRqs))
	r.currentRqs = newRqs
	return nil
}

type RedisRateLimiter struct {
	client     *redis.Client
	rqs        int
	tenantId   string
	workerName string
}

func NewRedisRateLimiter(workerName string, tid string, addr string, rqs int) *RedisRateLimiter {
	return &RedisRateLimiter{
		client: redis.NewClient(&redis.Options{
			Addr:     addr,
			Password: "", // no password set
			DB:       0,  // use default DB
		}),
		rqs:        rqs,
		tenantId:   tid,
		workerName: workerName, //debug only
	}
}

func (r *RedisRateLimiter) Take(ctx context.Context, unit int) error {
	for {
		err := r.take(ctx, unit)
		if err == nil {
			return nil
		}
		if err != redis.TxFailedErr {
			return err
		}
		r := rand.Float32()
		time.Sleep(time.Millisecond * time.Duration(r*100))
	}
}

func (r *RedisRateLimiter) take(ctx context.Context, unit int) error {
	if unit <= 0 || unit > int(r.rqs) {
		return &InvalidUnitError{}
	}

	bucketKey := r.tenantId + "_bucket"
	tsKey := r.tenantId + "_refillTs"

	allowed := false

	err := r.client.Watch(ctx, func(tx *redis.Tx) error {
		allowance, err := tx.Get(ctx, bucketKey).Float64()
		if err != nil {
			if err != redis.Nil {
				return err
			}
			allowance = float64(r.rqs)
		}
		refillTs, err := tx.Get(ctx, tsKey).Int64()
		if err != nil {
			if err != redis.Nil {
				return err
			}
			refillTs = time.Now().UnixNano()
		}

		//calculate new allowance
		currTs := time.Now()
		elapsed := currTs.UnixNano() - refillTs
		allowance += float64(elapsed) * float64(r.rqs) / float64(time.Second)

		//allowance cannot be bigger than quota
		if allowance > float64(r.rqs) {
			allowance = float64(r.rqs)
		}

		_, err = tx.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
			//fmt.Printf("(%s) %s allowance %v unit %d\n", r.workerName, ts(time.Now()), allowance, unit)
			if int(allowance) < unit {
				//not enough allowance ... cannot take
			} else {
				allowance -= float64(unit)
				allowed = true

				pipe.Set(ctx, bucketKey, strconv.FormatFloat(allowance, 'f', -1, 64), 0)
				pipe.Set(ctx, tsKey, strconv.FormatInt(currTs.UnixNano(), 10), 0)
			}
			return nil
		})
		return err
	}, bucketKey, tsKey)

	if err != nil {
		return err
	}
	if !allowed {
		return &LimitReached{}
	}
	return nil
}

type InvalidUnitError struct {
}

func (e *InvalidUnitError) Error() string {
	return "Invalid unit"
}

type LimitReached struct {
}

func (e *LimitReached) Error() string {
	return "Limit reached"
}

type BackendError struct {
	source error
}

func (e *BackendError) Error() string {
	return "BackendError: " + e.source.Error()
}

func (e *BackendError) Unwrap() error {
	return e.source
}
