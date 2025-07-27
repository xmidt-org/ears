// SPDX-FileCopyrightText: 2025 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package hasher

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
)

func Hash(i interface{}) string {
	switch v := i.(type) {
	default:
		return String(fmt.Sprintf("%v", v))
	}

}

func String(s string) string {
	hash := md5.Sum([]byte(s))
	return hex.EncodeToString(hash[:])
}
