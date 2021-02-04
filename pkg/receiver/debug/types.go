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
	"sync"

	"github.com/xmidt-org/ears/pkg/ledger"
	"github.com/xmidt-org/ears/pkg/validation"

	"github.com/xmidt-org/ears/pkg/receiver"

	"github.com/xorcare/pointer"
)

var _ receiver.Receiver = (*Receiver)(nil)

var DefaultConfig = &Config{
	IntervalMs: pointer.Int(100),
	Rounds:     pointer.Int(4),
	Payload:    "debug message",
	MaxHistory: pointer.Int(100),
}

var InfiniteRounds = pointer.Int(-1)

var _ validation.Validator = (*Config)(nil)
var _ validation.SchemaProvider = (*Config)(nil)

// Config determines how the receiver will operate.  To
// have the debug receiver run continuously, set the `Rounds` value
// to -1.
//
type Config struct {
	IntervalMs *int        `json:"intervalMs,omitempty"`
	Rounds     *int        `json:"rounds,omitempty"` // InfiniteRounds (-1) Signifies "infinite"
	Payload    interface{} `json:"payload,omitempty"`
	MaxHistory *int        `json:"maxHistory,omitempty"`
}

type Receiver struct {
	sync.Mutex
	done chan struct{}

	config  *Config
	history ledger.LimitedRecorder
	next    receiver.NextFn
}
