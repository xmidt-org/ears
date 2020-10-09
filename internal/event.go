/**
 *  Copyright (c) 2020  Comcast Cable Communications Management, LLC
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package internal

import (
	"context"
	"crypto/rand"
	"fmt"
)

type (
	// An Event bundles even payload and metadata that travels with the event
	Event struct {
		ctx      context.Context
		Payload  interface{} `json:"payload"`  // event payload (could also be string, but preparsed i sprobably more efficient if we want to allow deep inspection)
		Encoding string      `json:"encoding"` // optional encoding hint to be set by input plugin
		Metadata interface{} `json:"metadata"` // optional metadata produced by filter chain
		AckTree  AckTree     `json:"ackTree"`  // optional ack chain (or ack tree)
		Source   Pluginer    `json:"source"`   // pointer to source plugin instance
		TxId     string      `json:"txid"`     // transaction ID (probably also available from context)
		Ts       int         `json: "ts"`      // timestamp when event was received
	}
)

func NewEvent(ctx context.Context, plugin Pluginer, payload interface{}) *Event {
	evt := new(Event)
	evt.ctx = ctx
	evt.Payload = payload
	evt.Encoding = EncodingJson
	buf := make([]byte, 16)
	rand.Read(buf)
	evt.TxId = fmt.Sprintf("%x-%x-%x-%x-%x", buf[0:4], buf[4:6], buf[6:8], buf[8:10], buf[10:])
	evt.Source = plugin
	evt.Metadata = make(map[string]interface{})
	return evt
}
