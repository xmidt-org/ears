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

package main

import (
	"context"

	"github.com/xmidt-org/ears/pkg/receiver"
)

var _ receiver.Receiver = (*generator)(nil)

type generator struct {
	config string
}

func (g *generator) NewReceiver(config string) (receiver.Receiver, error) {
	g.config = config
}

func (g *generator) Receive(ctx context.Context, next receiver.NextFn) error {
	return nil
}
