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

package redis

import (
	"context"
	"github.com/go-redis/redis/v8"
	"github.com/pkg/errors"
	"github.com/xmidt-org/ears/pkg/ratelimit"
	"github.com/xmidt-org/ears/pkg/tenant"
	"math/rand"
	"strconv"
	"time"
)

type RedisRateLimiter struct {
	client *redis.Client
	rqs    int
	tid    tenant.Id
}

func NewRedisRateLimiter(tid tenant.Id, addr string, rqs int) *RedisRateLimiter {
	return &RedisRateLimiter{
		client: redis.NewClient(&redis.Options{
			Addr:     addr,
			Password: "", // no password set
			DB:       0,  // use default DB
		}),
		rqs: rqs,
		tid: tid,
	}
}

func (r *RedisRateLimiter) Limit() int {
	return r.rqs
}

func (r *RedisRateLimiter) SetLimit(newLimit int) error {
	if newLimit < 0 {
		return &ratelimit.InvalidUnitError{BadUnit: newLimit}
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
		if errors.Is(err, redis.TxFailedErr) {
			return err
		}
		r := rand.Float32()
		time.Sleep(time.Millisecond * time.Duration(r*100))
	}
}

func (r *RedisRateLimiter) take(ctx context.Context, unit int) error {
	if unit <= 0 || unit > r.rqs {
		return &ratelimit.InvalidUnitError{}
	}

	bucketKey := r.tid.Key() + "_bucket"
	tsKey := r.tid.Key() + "_refillTs"

	allowed := false

	err := r.client.Watch(ctx, func(tx *redis.Tx) error {
		allowance, err := tx.Get(ctx, bucketKey).Float64()
		if err != nil {
			if errors.Is(err, redis.Nil) {
				return err
			}
			allowance = float64(r.rqs)
		}
		refillTs, err := tx.Get(ctx, tsKey).Int64()
		if err != nil {
			if errors.Is(err, redis.Nil) {
				return err
			}
			refillTs = time.Now().UnixNano()
		}

		//calculate new allowance
		currTs := time.Now()
		elapsed := currTs.UnixNano() - refillTs

		//fmt.Printf("allowance=%f elapsed=%d\n", allowance, elapsed)

		allowance += float64(elapsed) * float64(r.rqs) / float64(time.Second)

		//allowance cannot be bigger than quota
		if allowance > float64(r.rqs) {
			allowance = float64(r.rqs)
		}

		_, err = tx.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
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
		return &ratelimit.LimitReached{}
	}
	return nil
}
