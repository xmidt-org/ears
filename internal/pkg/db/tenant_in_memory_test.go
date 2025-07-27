// SPDX-FileCopyrightText: 2025 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

//go:build !integration
// +build !integration

package db_test

import (
	"testing"

	"github.com/xmidt-org/ears/internal/pkg/db"
)

func TestInMemoryTenantStorer(t *testing.T) {
	s := db.NewTenantInmemoryStorer()
	testTenantStorer(s, t)
}
