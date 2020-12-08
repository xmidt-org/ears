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
	"io"
	"sync"

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

var Plugin = &plugin{}

type plugin struct {
	name    string
	version string
	config  string
}

type Receiver struct {
	IntervalMs int
	Rounds     int
	Payload    interface{}

	sync.Mutex
	done chan struct{}
}

type Sender struct {
	Destination io.Writer
}
