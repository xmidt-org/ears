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

	"github.com/xmidt-org/ears/pkg/event"
	"github.com/xmidt-org/ears/pkg/ledger"
	"github.com/xmidt-org/ears/pkg/validation"

	"github.com/xmidt-org/ears/pkg/sender"

	"github.com/xorcare/pointer"
)

var _ sender.Sender = (*Sender)(nil)

type EventWriter interface {
	Write(e event.Event) error
}

var _ EventWriter = (*SendStdout)(nil)

type SendStdout struct{}

var _ EventWriter = (*SendStdout)(nil)

type SendStderr struct{}

var _ EventWriter = (*SendSlice)(nil)

// SendSlice is an EventWriter that will store all the events in a slice.
// The data structure is unbounded, so make sure your debugging will complete
// after some expected amount of usage.  For long running debug sending,
// make use of Config.MaxHistory and History() instead.
type SendSlice struct {
	sync.Mutex
	events []event.Event
}

var DefaultConfig = &Config{
	Destination: DestinationDevNull,
	MaxHistory:  pointer.Int(0),
	Writer:      nil,
}

//go:generate rm -f destinationtype_enum.go
//go:generate go-enum -type=DestinationType -linecomment -sql=false
type DestinationType int

const (
	DestinationUnknown DestinationType = iota // unknown
	DestinationDevNull                        // devnull
	DestinationStdout                         // stdout
	DestinationStderr                         // stderr
	DestinationCustom                         // custom

)

var _ validation.Validator = (*Config)(nil)
var _ validation.SchemaProvider = (*Config)(nil)

// Config can be passed into NewSender() in order to configure
// the behavior of the sender.
type Config struct {

	// Destination should be set to one of the numerated DestinationType values.
	Destination DestinationType `json:"destination,omitempty"`

	// MaxHistory defines the number of past events to keep.  Begins replacing the
	// oldest event when the buffer is full.
	MaxHistory *int `json:"maxHistory,omitempty"`

	// Writer defines a custom writer that will be written to on each Send.
	// Writer should support concurrent writes.  NOTE: We do not support
	// (Un)Marshalling of the value as it's an interface pointer.
	//
	Writer EventWriter `json:"-"`
}

type Sender struct {
	sync.Mutex

	config  *Config
	history ledger.LimitedRecorder

	destination EventWriter
}
