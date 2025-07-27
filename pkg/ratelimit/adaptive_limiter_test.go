// SPDX-FileCopyrightText: 2025 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package ratelimit_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/xmidt-org/ears/pkg/ratelimit"
	"github.com/xmidt-org/ears/pkg/tenant"
)

func TestAdaptiveRateLimiter(t *testing.T) {
	backend := ratelimit.NewInMemoryBackendLimiter(tenant.Id{OrgId: "myApp", AppId: "myOrg"}, 12)
	limiter := ratelimit.NewAdaptiveRateLimiter(backend, 1, 12)

	ctx := context.Background()
	logger := log.With().Logger()
	ctx = logger.WithContext(ctx)

	subCtx, cancel := context.WithTimeout(ctx, time.Second*5)
	err := simulateRps(limiter, 1, subCtx)
	if err != nil {
		t.Fatalf("Failed to simulateRps %s\n", err.Error())
	}
	cancel()

	//verify that we can take at the desired rate
	err = validateRps(limiter, 1, ctx)
	if err != nil {
		t.Fatalf("Fail to validate at 1 rps, error=%s\n", err.Error())
	}

	subCtx, cancel = context.WithTimeout(ctx, time.Second*5)
	err = simulateRps(limiter, 4, subCtx)
	if err != nil {
		t.Fatalf("Failed to simulateRps %s\n", err.Error())
	}
	cancel()

	//verify that we can take at the desired rate
	err = validateRps(limiter, 4, ctx)
	if err != nil {
		t.Fatalf("Fail to validate at 4 rps, error=%s\n", err.Error())
	}

	subCtx, cancel = context.WithTimeout(ctx, time.Second*5)
	err = simulateRps(limiter, 9, subCtx)
	if err != nil {
		t.Fatalf("Failed to simulateRps %s\n", err.Error())
	}
	cancel()

	//verify that we can take at the desired rate
	err = validateRps(limiter, 9, ctx)
	if err != nil {
		t.Fatalf("Fail to validate at 9 rps, error=%s\n", err.Error())
	}

	subCtx, cancel = context.WithTimeout(ctx, time.Second*5)
	err = simulateRps(limiter, 3, subCtx)
	if err != nil {
		t.Fatalf("Failed to simulateRps %s\n", err.Error())
	}
	cancel()

	//verify that we can take at the desired rate
	err = validateRps(limiter, 3, ctx)
	if err != nil {
		t.Fatalf("Fail to validate at 3 rps, error=%s\n", err.Error())
	}

	subCtx, cancel = context.WithTimeout(ctx, time.Second*5)
	err = simulateRps(limiter, 15, subCtx)
	if err == nil {
		t.Fatalf("Expect an error that limit won't converge. Instead, got no error")
	}
	cancel()
}

func validateRps(limiter *ratelimit.AdaptiveRateLimiter, rps int, ctx context.Context) error {
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

func simulateRps(limiter *ratelimit.AdaptiveRateLimiter, rps int, ctx context.Context) error {
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
