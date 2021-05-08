package ratelimit

import (
	"context"
	"github.com/go-redis/redis/v8"
	"math/rand"
	"strconv"
	"time"
)

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

func (r *RedisRateLimiter) Limit() int {
	return r.rqs
}

func (r *RedisRateLimiter) SetLimit(newLimit int) error {
	if newLimit < 0 {
		return &InvalidUnitError{newLimit}
	}
	r.rqs = newLimit
	return nil
}

func (r *RedisRateLimiter) Take(ctx context.Context, unit int) error {
	if r.rqs == 0 {
		//Unlimited
		return nil
	}

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
