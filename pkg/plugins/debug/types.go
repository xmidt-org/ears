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
	"container/ring"
	"sync"

	"github.com/xmidt-org/ears/pkg/event"
	pkgplugin "github.com/xmidt-org/ears/pkg/plugin"

	"github.com/xmidt-org/ears/pkg/receiver"
	"github.com/xmidt-org/ears/pkg/sender"

	"github.com/xorcare/pointer"
)

var _ pkgplugin.Pluginer = (*plugin)(nil)
var _ pkgplugin.NewPluginerer = (*plugin)(nil)
var _ receiver.NewReceiverer = (*plugin)(nil)
var _ sender.NewSenderer = (*plugin)(nil)

var _ sender.Sender = (*Sender)(nil)
var _ receiver.Receiver = (*Receiver)(nil)

var (
	Name    = "debug"
	Version = "v0.0.0"
	Commit  = ""
)

func NewPlugin() *plugin {
	return NewPluginVersion(Name, Version, Commit)
}

func NewPluginVersion(name string, version string, commit string) *plugin {
	return &plugin{
		name:    name,
		version: version,
		commit:  commit,
	}
}

type plugin struct {
	name    string
	version string
	commit  string
	config  interface{}
}

var DefaultReceiverConfig = ReceiverConfig{
	IntervalMs: pointer.Int(100),
	Rounds:     pointer.Int(4),
	Payload:    "debug message",
	MaxHistory: pointer.Int(100),
}

// ReceiverConfig determines how the receiver will operate.  To
// have the debug receiver run continuously, set the `Rounds` value
// to -1.
//

var InfiniteRounds = pointer.Int(-1)

type ReceiverConfig struct {
	IntervalMs *int        `json:"intervalMs,omitempty"`
	Rounds     *int        `json:"rounds,omitempty"` // InfiniteRounds (-1) Signifies "infinite"
	Payload    interface{} `json:"payload,omitempty"`
	MaxHistory *int        `json:"maxHistory,omitempty"`
}

type Receiver struct {
	sync.Mutex
	done chan struct{}

	config  ReceiverConfig
	history *history
	next    receiver.NextFn
}

type EventWriter interface {
	Write(e event.Event) error
}

type SendStdout struct {
	EventWriter
}

type SendStderr struct {
	EventWriter
}

// SendSlice is an EventWriter that will store all the events in a slice.
// The data structure is unbounded, so make sure your debugging will complete
// after some expected amount of usage.  For long running debug sending,
// make use of SenderConfig.MaxHistory and History() instead.
type SendSlice struct {
	EventWriter

	sync.Mutex
	events []event.Event
}

var minSenderConfig = SenderConfig{
	MaxHistory: pointer.Int(0),
}

var DefaultSenderConfig = SenderConfig{
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

// SenderConfig can be passed into NewSender() in order to configure
// the behavior of the sender.
type SenderConfig struct {

	// Destination should be set to one of the numerated DestinationType values.
	Destination DestinationType `json:"destination,omitempty"`

	// MaxHistory defines the number of past events to keep.  Begins replacing the
	// oldest event when the buffer is full.
	MaxHistory *int `json:"maxHistory,omitempty"`

	// Writer defines a custom writer that will be written to on each Send.
	// Writer should support concurrent writes.
	Writer EventWriter `json:"-"`
}

type Sender struct {
	sync.Mutex

	config  SenderConfig
	history *history

	destination EventWriter
}

type history struct {
	sync.Mutex

	size  int
	count int
	ring  *ring.Ring
}
