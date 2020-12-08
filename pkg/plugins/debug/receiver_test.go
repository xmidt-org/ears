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
	"fmt"

	"testing"
	"time"

	"github.com/xmidt-org/ears/pkg/plugins/debug"

	"github.com/xmidt-org/ears/pkg/event"
)

func TestReceiver(t *testing.T) {
	r, err := debug.NewReceiver("")
	if err != nil {
		t.Errorf(err.Error())
	}

	ctx, cancel := context.WithTimeout(context.Background(), 12*time.Second)
	defer cancel()

	err = r.Receive(ctx, func(ctx context.Context, e event.Event) error {
		fmt.Printf("was sent a message: %+v\n", e)
		return nil
	})

	if err != nil {
		fmt.Println("receive returned an error: ", err)
	}
}
