// SPDX-FileCopyrightText: 2025 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package debug

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"
)

// WithDefaults will set
//   - The .Destination and .Writer if Destination is DestinationUnknown
//   - The .MaxHistory if it is nil
//
// The values will be set to the corresponding DefaultSenderConfig.* values
func (sc SenderConfig) WithDefaults() SenderConfig {
	cfg := sc

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
//   - The .Destination is a valid DestinationType and is not DestinationUnknown
//   - Max history is >= 0
//   - The .Writer is only set when .Destination is DestinationCustom
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
