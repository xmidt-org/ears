package main

import (
	"context"
	"errors"
	"fmt"
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

	//for i := 0; i < numWorker; i++ {
	go worker("0", &wg, 6, 30, 4, 20)
	go worker("1", &wg, 6, 30, 4, 20)
	go worker("2", &wg, 6, 30, 4, 20)
	go worker("3", &wg, 6, 30, 4, 20)
	go worker("4", &wg, 6, 30, 4, 20)
	//}
	wg.Wait()

	fmt.Printf("success %d, fail %d, error %d\n", successCount, failCount, errCount)
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
	backendLimiter := NewRedisRateLimiter(workerName, "mchiang", redisAddr, maxRps)
	limiter := NewAdaptiveRateLimiter(workerName, backendLimiter, initialRps, maxRps)

	ctx := context.Background()

	total := rps * second
	for i := 0; i < total; i++ {
		start := time.Now()
		err := limiter.Take(ctx, 1)
		end := time.Now()

		fmt.Printf("(%s) start %s, end %s, (%s)\n", workerName, ts(start), ts(end), errStr(err))
		if err != nil {
			var limitReachErr *LimitReached
			if errors.As(err, &limitReachErr) {
				atomic.AddInt32(&failCount, 1)
			} else {
				fmt.Printf("error %s\n", err.Error())
				atomic.AddInt32(&errCount, 1)
			}
			time.Sleep(time.Second / time.Duration(rps))
		} else {
			time.Sleep(time.Second / time.Duration(rps))
			atomic.AddInt32(&successCount, 1)
		}
	}
}
