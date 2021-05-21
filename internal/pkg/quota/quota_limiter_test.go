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

package quota_test

import (
	"context"
	"errors"
	"fmt"
	"github.com/xmidt-org/ears/internal/pkg/quota"
	"github.com/xmidt-org/ears/pkg/tenant"
	"testing"
	"time"
)

func TestQuotaLimiter(t *testing.T) {
	limiter := quota.NewQuotaLimiter(tenant.Id{OrgId: "myOrg", AppId: "myApp"}, "inmemory", "", 1, 12)

	var err error
	for i := 0; i < 3; i++ {
		//validate that we can do 1rps
		err = validateRps(limiter, 1)
		if err == nil {
			break
		}
		if errors.Is(err, TestErr_FailToReachRps) {
			continue
		}
		t.Fatalf("Fail to reach the desired rps, error=%s\n", err.Error())
	}
	if err != nil {
		t.Fatalf("Fail to reach the desired rps in 3 tries")
	}

	for i := 0; i < 3; i++ {
		//validate that we can do 2rps
		err = validateRps(limiter, 2)
		if err == nil {
			break
		}
		if errors.Is(err, TestErr_FailToReachRps) {
			time.Sleep(time.Second)
			continue
		}
		t.Fatalf("Fail to reach the desired rps, error=%s\n", err.Error())
	}
	if err != nil {
		t.Fatalf("Fail to reach the desired rps in 3 tries")
	}

	for i := 0; i < 3; i++ {
		//validate that we can do 8rps
		err = validateRps(limiter, 8)
		if err == nil {
			break
		}
		if errors.Is(err, TestErr_FailToReachRps) {
			continue
		}
		t.Fatalf("Fail to reach the desired rps, error=%s\n", err.Error())
	}
	if err != nil {
		t.Fatalf("Fail to reach the desired rps in 3 tries")
	}

	limiter.SetLimit(100)
	for i := 0; i < 10; i++ {
		//validate that we can do 80rps
		err = validateRps(limiter, 80)
		if err == nil {
			break
		}
		if err.Error() == "Cannot reach desired RPS" {
			continue
		}
		t.Fatalf("Fail to reach the desired rps, error=%s\n", err.Error())
	}
	if err != nil {
		t.Fatalf("Fail to reach the desired rps in 3 tries")
	}

	limiter.SetLimit(1000)
	for i := 0; i < 10; i++ {
		//validate that we can do 80rps
		err = validateRps(limiter, 800)
		if err == nil {
			break
		}
		if err.Error() == "Cannot reach desired RPS" {
			continue
		}
		t.Fatalf("Fail to reach the desired rps, error=%s\n", err.Error())
	}
	if err != nil {
		t.Fatalf("Fail to reach the desired rps in 3 tries")
	}
}

func validateRps(limiter *quota.QuotaLimiter, rps int) error {
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	start := time.Now()

	count := 0
	for i := 0; i < rps; i++ {
		err := limiter.Wait(ctx)
		if err != nil {
			return err
		}
		count++
	}

	fmt.Printf("elapsed time=%dms, adaptiveLimit=%d\n", time.Since(start).Milliseconds(), limiter.AdaptiveLimit())
	if start.Add(time.Second + time.Millisecond*100).After(time.Now()) {
		return nil
	}
	return TestErr_FailToReachRps
}
