// SPDX-FileCopyrightText: 2025 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package gears

import (
	"fmt"
	"hash/fnv"
	"testing"
)

func TestLocationHash(t *testing.T) {
	location := "MChian100"
	hashbuf := []byte(location)
	h := fnv.New32a()
	h.Write(hashbuf)
	pIdx := getProducerIdx(h.Sum32(), 2)
	fmt.Printf("index=%d\n", pIdx)
}
