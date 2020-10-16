// Licensed to Comcast Cable Communications Management, LLC under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. Comcast Cable Communications Management, LLC licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package sender

import (
	"context"

	"github.com/xmidt-org/ears/pkg/event"
)

type NewSenderer interface {
	NewSender(config string) (Sender, error)
}

// or Outputter[√] or Producer[x] or Publisher[√]
type Sender interface {
	Hash() string // SenderHash ?
	Send(ctx context.Context, e event.Event) error
}
