// SPDX-FileCopyrightText: 2025 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package nop

func (sc SenderConfig) WithDefaults() SenderConfig {
	cfg := sc
	return cfg
}

func (sc *SenderConfig) Validate() error {
	return nil
}
