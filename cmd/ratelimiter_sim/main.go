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

package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/rs/zerolog/log"
	"github.com/xmidt-org/ears/pkg/ratelimit"
	"github.com/xmidt-org/ears/pkg/tenant"
	"sync"
	"sync/atomic"
	"time"
)

const redisAddr = "127.0.0.1:6379"

var successCount int32 = 0
var failCount int32 = 0
var errCount int32 = 0

func main() {
	wg := sync.WaitGroup{}
	numWorker := 5
	wg.Add(numWorker)

	duration := 20
	quota := 1000

	go worker("0", &wg, 300, duration, quota/numWorker, quota)
	go worker("1", &wg, 300, duration, quota/numWorker, quota)
	go worker("2", &wg, 100, duration, quota/numWorker, quota)
	go worker("3", &wg, 100, duration, quota/numWorker, quota)
	go worker("4", &wg, 20, duration, quota/numWorker, quota)

	reporter(duration + 10)
	wg.Wait()

	fmt.Printf("Total success %d, fail %d, error %d\n", successCount, failCount, errCount)
}

func ts(t time.Time) string {
	return t.Format("04:05.000")
}

func errStr(err error) string {
	if err == nil {
		return "success"
	}
	return err.Error()
}

func worker(workerName string, wg *sync.WaitGroup, rps int, second int, initialRps, maxRps int) {
	defer wg.Done()
	backendLimiter := ratelimit.NewRedisRateLimiter(tenant.Id{"myOrg", "mchiang"}, redisAddr, maxRps)
	limiter := ratelimit.NewAdaptiveRateLimiter(backendLimiter, initialRps, maxRps)

	ctx := context.Background()
	logger := log.With().Str("workerName", workerName).Int("desired", rps).Logger()
	ctx = logger.WithContext(ctx)

	total := rps * second
	for i := 0; i < total; i++ {
		start := time.Now()
		err := limiter.Take(ctx, 1)
		end := time.Now()

		//fmt.Printf("(%s) start %s, end %s, (%s)\n", workerName, ts(start), ts(end), errStr(err))
		if err != nil {
			//fmt.Printf("(%s) start %s, end %s, (%s)\n", workerName, ts(start), ts(end), errStr(err))
			var limitReachErr *ratelimit.LimitReached
			if errors.As(err, &limitReachErr) {
				atomic.AddInt32(&failCount, 1)
			} else {
				fmt.Printf("(%s) start %s, end %s, (%s)\n", workerName, ts(start), ts(end), errStr(err))
				atomic.AddInt32(&errCount, 1)
			}
			time.Sleep(time.Second / time.Duration(rps))
		} else {
			time.Sleep(time.Second / time.Duration(rps))
			atomic.AddInt32(&successCount, 1)
		}
	}
}

func reporter(duration int) {
	var sCount int32 = 0
	var fCount int32 = 0
	var eCount int32 = 0
	for i := 0; i < duration; i++ {
		time.Sleep(time.Second)
		sTmp := successCount
		fTmp := failCount
		eTmp := errCount
		fmt.Printf("success %d, fail %d, error %d\n", sTmp-sCount, fTmp-fCount, eTmp-eCount)
		sCount = sTmp
		fCount = fTmp
		eCount = eTmp
	}
}
