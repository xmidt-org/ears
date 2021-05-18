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

package ratelimit_test

import (
	"context"
	"github.com/pkg/errors"
	"github.com/xmidt-org/ears/pkg/ratelimit"
	"testing"
	"time"
)

func testBackendLimiter(r ratelimit.RateLimiter, t *testing.T) {

	err := r.SetLimit(-1)
	var invalidUnit *ratelimit.InvalidUnitError
	if err == nil {
		t.Fatalf("Expect InvalidUnitError, got no error instead")
	}
	if !errors.As(err, &invalidUnit) {
		t.Fatalf("Expecting InvalidUnitError, got %s instead", err.Error())
	}

	err = r.SetLimit(5)
	if err != nil {
		t.Fatalf("Fail to set new limit %s\n", err.Error())
	}

	ctx := context.Background()

	time.Sleep(time.Second)
	err = r.Take(ctx, 5)
	if err != nil {
		t.Fatalf("Fail to take 5 unit, error=%s\n", err.Error())
	}

	for i := 0; i < 5; i++ {
		time.Sleep(200 * time.Millisecond)
		err = r.Take(ctx, 1)
		if err != nil {
			t.Fatalf("Fail to take on %d, error=%s\n", i, err.Error())
		}
	}

	//next one should fail
	err = r.Take(ctx, 1)
	if err == nil {
		t.Fatalf("Expect LimitReached, got no error instead")
	}
	var limitReached *ratelimit.LimitReached
	if !errors.As(err, &limitReached) {
		t.Fatalf("Expect LimitReached, got %s instead", err.Error())
	}

	//try taking at 2x the rate
	for i := 0; i < 5; i++ {
		time.Sleep(100 * time.Millisecond)
		err = r.Take(ctx, 1)
		if err == nil {
			t.Fatalf("Expect LimitReached, got no error instead")
		}
		if !errors.As(err, &limitReached) {
			t.Fatalf("Expect LimitReached, got %s instead", err.Error())
		}
		time.Sleep(100 * time.Millisecond)
		err = r.Take(ctx, 1)
		if err != nil {
			t.Fatalf("Fail to take on %d, error=%s\n", i, err.Error())
		}
	}

	r.SetLimit(10)
	for i := 0; i < 10; i++ {
		time.Sleep(100 * time.Millisecond)
		err = r.Take(ctx, 1)
		if err != nil {
			t.Fatalf("Fail to take on %d, error=%s\n", i, err.Error())
		}
	}

	//next one should fail
	err = r.Take(ctx, 1)
	if err == nil {
		t.Fatalf("Expect LimitReached, got no error instead")
	}
	if !errors.As(err, &limitReached) {
		t.Fatalf("Expect LimitReached, got %s instead", err.Error())
	}
}
