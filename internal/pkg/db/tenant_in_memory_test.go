// +build !integration

package db_test

import (
	"github.com/xmidt-org/ears/internal/pkg/db"
	"testing"
)

func TestInMemoryTenantStorer(t *testing.T) {
	s := db.NewTenantInmemoryStorer()
	testTenantStorer(s, t)
}
