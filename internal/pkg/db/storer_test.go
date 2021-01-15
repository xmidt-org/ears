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

package db_test

import (
	"context"
	"testing"

	"github.com/xmidt-org/ears/internal/pkg/db"
	"github.com/xmidt-org/ears/pkg/route"
)

func TestInMemoryRouteStorer(t *testing.T) {
	s := db.NewInMemoryRouteStorer(nil)

	testRouteStorer(s, t)
}

func testRouteStorer(s route.RouteStorer, t *testing.T) {
	ctx := context.Background()

	//start from a clean slate
	err := s.DeleteAllRoutes(ctx)
	if err != nil {
		t.Fatalf("DeleteAllRoutes error: %s\n", err.Error())
	}

	r, err := s.GetRoute(ctx, "does_not_exist")
	if err != nil {
		t.Fatalf("GetRoute error: %s\n", err.Error())
	}
	if r != nil {
		t.Fatalf("Expect empty route but instead get a route")
	}
}
