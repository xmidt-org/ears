// SPDX-FileCopyrightText: 2025 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"errors"
	"testing"
)

func TestErrors(t *testing.T) {
	err := &InvalidOptionError{
		Option: "bad option",
	}
	var errType *InvalidOptionError
	if !errors.As(err, &errType) {
		t.Error("Error comparison failed")
	}

}
