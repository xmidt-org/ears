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

func NewPlugin(name string, version string, commit string) *plugin {
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

var defaultReceiverConfig = ReceiverConfig{
	IntervalMs: 100,
	Rounds:     4,
	Payload:    "debug message",
	MaxHistory: 100,
}

type ReceiverConfig struct {
	IntervalMs int         `json:"intervalMs"`
	Rounds     int         `json:"rounds"` // -1 Signifies "infinite"
	Payload    interface{} `json:"payload"`
	MaxHistory int         `json:"maxHistory"`
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

type SenderConfig struct {
	Destination string      `json:"destination"` // "", stdout, stderr
	MaxHistory  int         `json:"maxHistory"`
	Writer      EventWriter `json:"-"`
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
