// SPDX-FileCopyrightText: 2025 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

//go:build integration
// +build integration

package ratelimit_test

import (
	"testing"

	"github.com/xmidt-org/ears/pkg/ratelimit/redis"
	"github.com/xmidt-org/ears/pkg/tenant"
)

func TestRedisBackendLimiter(t *testing.T) {
	limiter := redis.NewRedisRateLimiter(
		tenant.Id{"myOrg", "myUnitTestApp"},
		"localhost:6379",
		0,
	)

	testBackendLimiter(limiter, t)
}
