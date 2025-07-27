// SPDX-FileCopyrightText: 2025 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package ratelimit_test

import (
	"testing"

	"github.com/xmidt-org/ears/pkg/ratelimit"
	"github.com/xmidt-org/ears/pkg/tenant"
)

func TestInMemoryBackendLimiter(t *testing.T) {
	limiter := ratelimit.NewInMemoryBackendLimiter(tenant.Id{OrgId: "myOrg", AppId: "myApp"}, 0)

	testBackendLimiter(limiter, t)
}
