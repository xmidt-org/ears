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

package route

import (
	"context"
	"sync"

	"github.com/xmidt-org/ears/pkg/filter"
	"github.com/xmidt-org/ears/pkg/receiver"
	"github.com/xmidt-org/ears/pkg/sender"
)

type Router interface {
	Run(ctx context.Context, r receiver.Receiver, f filter.Filterer, s sender.Sender) error
	Stop(ctx context.Context) error
}

type Route struct {
	sync.Mutex

	r receiver.Receiver
	f filter.Filterer
	s sender.Sender

	done chan struct{}
}

type InvalidRouteError struct {
	Err error
}
