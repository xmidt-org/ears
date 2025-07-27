// SPDX-FileCopyrightText: 2025 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package nop

func (rc *ReceiverConfig) WithDefaults() ReceiverConfig {
	cfg := *rc
	return cfg
}

// Validate returns an error upon validation failure
func (rc *ReceiverConfig) Validate() error {
	return nil
}
