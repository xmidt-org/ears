// SPDX-FileCopyrightText: 2025 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

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
