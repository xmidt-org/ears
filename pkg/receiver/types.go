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

package receiver

import (
	"context"

	"github.com/xmidt-org/ears/pkg/event"
)

type NewReceiverer interface {
	NewReceiver(config string) (Receiver, error)
}

/*
Receiver is a plugin that will receive messages and will
send them to `NextFn`

Other discarded names:
* Inputter[x]
* Ingestor[x]
* Consumer[x]
* Subscriber[√]
* Listener[?]  vs Speaker?!?!?

*/
type Receiver interface {
	Hash() string // ReceiverHash ?

	Receive(ctx context.Context, next func(event.Event)) error
}
