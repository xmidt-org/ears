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
	"errors"
	"github.com/rs/zerolog/log"
	"github.com/xmidt-org/ears/pkg/tenant"
	"testing"
	"time"
)

func TestAdaptiveRateLimiter(t *testing.T) {
	backend := NewInMemoryBackendLimiter(tenant.Id{"myApp", "myOrg"}, 12)
	limiter := NewAdaptiveRateLimiter(backend, 1, 12)

	ctx := context.Background()
	logger := log.With().Logger()
	ctx = logger.WithContext(ctx)

	subCtx, _ := context.WithTimeout(ctx, time.Second*5)
	err := simulateRps(limiter, 1, subCtx)
	if err != nil {
		t.Fatalf("Failed to simulateRps %s\n", err.Error())
	}
	//verify that we can take at the desired rate
	err = validateRps(limiter, 1, ctx)
	if err != nil {
		t.Fatalf("Fail to validate at 1 rps, error=%s\n", err.Error())
	}

	subCtx, _ = context.WithTimeout(ctx, time.Second*5)
	err = simulateRps(limiter, 4, subCtx)
	if err != nil {
		t.Fatalf("Failed to simulateRps %s\n", err.Error())
	}
	//verify that we can take at the desired rate
	err = validateRps(limiter, 4, ctx)
	if err != nil {
		t.Fatalf("Fail to validate at 4 rps, error=%s\n", err.Error())
	}

	subCtx, _ = context.WithTimeout(ctx, time.Second*5)
	err = simulateRps(limiter, 9, subCtx)
	if err != nil {
		t.Fatalf("Failed to simulateRps %s\n", err.Error())
	}
	//verify that we can take at the desired rate
	err = validateRps(limiter, 9, ctx)
	if err != nil {
		t.Fatalf("Fail to validate at 9 rps, error=%s\n", err.Error())
	}

	subCtx, _ = context.WithTimeout(ctx, time.Second*5)
	err = simulateRps(limiter, 3, subCtx)
	if err != nil {
		t.Fatalf("Failed to simulateRps %s\n", err.Error())
	}
	//verify that we can take at the desired rate
	err = validateRps(limiter, 3, ctx)
	if err != nil {
		t.Fatalf("Fail to validate at 3 rps, error=%s\n", err.Error())
	}

	subCtx, _ = context.WithTimeout(ctx, time.Second*5)
	err = simulateRps(limiter, 15, subCtx)
	if err == nil {
		t.Fatalf("Expect an error that limit won't converge. Instead, got no error")
	}
}

func validateRps(limiter *AdaptiveRateLimiter, rps int, ctx context.Context) error {
	sleepTime := time.Duration(1000000/rps) * time.Microsecond
	for i := 0; i < rps; i++ {
		time.Sleep(sleepTime)
		err := limiter.Take(ctx, 1)
		if err != nil {
			return err
		}
	}
	return nil
}

func simulateRps(limiter *AdaptiveRateLimiter, rps int, ctx context.Context) error {
	sleepTime := time.Duration(1000000/rps) * time.Microsecond
	prevLimit := limiter.AdaptiveLimit()
	for {
		select {
		case <-ctx.Done():
			//adaptive rate limiter cannot converged
			return errors.New("adapative rate limiter does not converge")
		case <-time.After(sleepTime):
		}
		limiter.Take(ctx, 1)

		if limiter.AdaptiveLimit() >= rps && limiter.AdaptiveLimit() == prevLimit {
			break
		}
		prevLimit = limiter.AdaptiveLimit()
	}
	return nil
}
