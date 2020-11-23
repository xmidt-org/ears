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

package debug_test

import (
	"context"

	"testing"
	"time"

	"github.com/xmidt-org/ears/pkg/plugins/debug"

	"github.com/xmidt-org/ears/pkg/event"
)

func TestSender(t *testing.T) {
	s, err := debug.NewSender("")
	if err != nil {
		t.Errorf(err.Error())
	}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	e, err := event.NewEvent("hello")
	if err != nil {
		t.Errorf(err.Error())
	}

	err = s.Send(ctx, e)

	if err != nil {
		t.Errorf(err.Error())
	}
}
