// SPDX-FileCopyrightText: 2025 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package debug_test

func min(vals ...int) int {
	if len(vals) < 1 {
		return 0
	}

	m := vals[0]
	for _, v := range vals {
		if v < m {
			m = v
		}
	}
	return m
}
