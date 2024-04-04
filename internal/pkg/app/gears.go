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

package app

// Gears Envelope

type GearsEnvelope struct {
	Message GearsEnvelopeMessage `json:"message,omitempty" xml:"message,omitempty"`
	To      GearsEnvelopeTo      `json:"to,omitempty" xml:"to,omitempty"`
	Tx      GearsEnvelopeTx      `json:"tx,omitempty" xml:"tx,omitempty"`
}

type GearsEnvelopeMessage struct {
	Op      string      `json:"op,omitempty" xml:"op,omitempty"`
	Payload interface{} `json:"payload,omitempty" xml:"payload,omitempty"`
}

type GearsEnvelopeTo struct {
	Partner  string `json:"partner,omitempty" xml:"partner,omitempty"`
	App      string `json:"app,omitempty" xml:"app,omitempty"`
	Location string `json:"location,omitempty" xml:"location,omitempty"`
}

type GearsEnvelopeTx struct {
	TraceId string `json:"trace.id,omitempty" xml:"trace.id,omitempty"`
}

func NewGearsEnvelope(partner, app, location, traceId string, payload interface{}) *GearsEnvelope {
	envelope := new(GearsEnvelope)
	envelope.Message.Payload = payload
	envelope.Message.Op = "process"
	if partner == "" {
		envelope.To.Partner = "comcast"
	} else {
		envelope.To.Partner = partner
	}
	envelope.To.App = app
	envelope.To.Location = location
	if traceId != "" {
		envelope.Tx.TraceId = traceId
	}
	return envelope
}
