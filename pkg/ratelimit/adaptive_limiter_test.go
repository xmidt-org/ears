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
