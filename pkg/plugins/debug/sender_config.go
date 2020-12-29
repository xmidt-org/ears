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

// WithDefaults will set
//   * The .Destination and .Writer if Destination is DestinationUnknown
//   * The .MaxHistory if it is nil
//
// The values will be set to the corresponding DefaultSenderConfig.* values
//
func (sc *SenderConfig) WithDefaults() SenderConfig {
	cfg := *sc

	if sc.Destination == DestinationUnknown {
		cfg.Destination = DefaultSenderConfig.Destination
		cfg.Writer = DefaultSenderConfig.Writer
	}

	if sc.MaxHistory == nil {
		cfg.MaxHistory = DefaultSenderConfig.MaxHistory
	}

	return cfg
}

// Validate will ensure that:
//   * The .Destination is a valid DestinationType and is not DestinationUnknown
//   * Max history is >= 0
//   * The .Writer is only set when .Destination is DestinationCustom
func (sc *SenderConfig) Validate() error {
	s := *sc

	// Allow this list to easily expand over time
	validDestinationTypes := []interface{}{}
	for _, t := range DestinationTypeValues() {
		if t != DestinationUnknown {
			validDestinationTypes = append(validDestinationTypes, t)
		}
	}

	return validation.ValidateStruct(&s,
		validation.Field(&s.Destination,
			validation.Required,
			validation.In(validDestinationTypes...),
		),

		validation.Field(&s.MaxHistory,
			validation.NotNil,
			validation.Min(*minSenderConfig.MaxHistory),
		),

		validation.Field(&s.Writer,
			validation.When(
				s.Destination != DestinationCustom,
				validation.Nil,
			),
			validation.When(
				s.Destination == DestinationCustom,
				validation.NotNil,
			),
		),
	)

}
