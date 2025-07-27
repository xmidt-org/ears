// SPDX-FileCopyrightText: 2025 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package bit

import "fmt"

type Mask uint

func (b *Mask) Set(flag Mask) {
	*b = *b | flag
}

func (b *Mask) Clear(flag Mask) {
	*b = *b & ^flag
}

func (b *Mask) Flip(flag Mask) {
	*b ^= flag
}

func (b Mask) IsSet(flag Mask) bool {
	return b&flag != 0
}

func (b Mask) String() string {
	return fmt.Sprintf("%b", b)
}
