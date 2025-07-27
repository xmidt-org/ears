//go:build race
// +build race

// https://stackoverflow.com/questions/44944959/how-can-i-check-if-the-race-detector-is-enabled-at-runtime

// SPDX-FileCopyrightText: 2025 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package manager_test

var raceFlagEnabled = true
