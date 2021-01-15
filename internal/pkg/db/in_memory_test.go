// +build !integration

package db_test

import (
	"github.com/xmidt-org/ears/internal/pkg/db"
	"testing"
)

func TestInMemoryRouteStorer(t *testing.T) {
	s := db.NewInMemoryRouteStorer(nil)
	testRouteStorer(s, t)
}
