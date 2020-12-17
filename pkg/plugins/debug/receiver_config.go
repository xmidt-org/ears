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
	validation "github.com/go-ozzo/ozzo-validation/v4"
)

var minReceiverConfig = ReceiverConfig{
	IntervalMs: 1,
	Rounds:     -1,
	MaxHistory: 0,
}

func (rc *ReceiverConfig) WithDefaults() *ReceiverConfig {
	cfg := *rc

	if cfg.IntervalMs <= 0 {
		cfg.IntervalMs = defaultReceiverConfig.IntervalMs
	}

	if cfg.Rounds < 0 {
		cfg.Rounds = defaultReceiverConfig.Rounds
	}

	if cfg.Payload == nil {
		cfg.Payload = defaultReceiverConfig.Payload
	}

	return &cfg
}

func (rc *ReceiverConfig) Validate() error {
	r := *rc
	return validation.ValidateStruct(&r,
		validation.Field(&r.IntervalMs, validation.Min(minReceiverConfig.IntervalMs)),
		validation.Field(&r.Rounds, validation.Min(minReceiverConfig.Rounds)),
		validation.Field(&r.Payload, validation.NotNil),
		validation.Field(&r.MaxHistory, validation.Min(minReceiverConfig.MaxHistory)),
	)

}
