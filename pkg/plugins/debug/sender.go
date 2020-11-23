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

package debug

import (
	"context"
	"fmt"
	"os"

	"github.com/xmidt-org/ears/pkg/event"
	"github.com/xmidt-org/ears/pkg/sender"
)

func NewSender(config string) (sender.Sender, error) {
	return &Sender{
		Destination: os.Stdout,
	}, nil
}

func (s *Sender) Send(ctx context.Context, e event.Event) error {
	if s.Destination == nil {
		return &sender.InvalidConfigError{
			Err: fmt.Errorf("destination cannot be null"),
		}
	}

	fmt.Fprintln(s.Destination, e)
	return nil
}
