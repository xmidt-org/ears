// SPDX-FileCopyrightText: 2025 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package panics_test

import (
	"testing"

	"github.com/xmidt-org/ears/pkg/panics"
)

func TestToError(t *testing.T) {
	myPanic := "my panic"

	defer func() {
		p := recover()
		err := panics.ToError(p)
		if err.Error() != myPanic {
			t.Errorf("wrong panic error")
		}
	}()
	panic(myPanic)
}
