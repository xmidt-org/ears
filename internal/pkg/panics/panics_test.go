// Copyright 2020 Comcast Cable Communications Management, LLC
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

package panics_test

import (
	"github.com/xmidt-org/ears/internal/pkg/panics"
	"testing"
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
