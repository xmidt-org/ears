// SPDX-FileCopyrightText: 2025 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package config

// Config interface for uber/fx
type Config interface {
	GetString(key string) string
	GetInt(key string) int
	GetBool(key string) bool
}
