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
