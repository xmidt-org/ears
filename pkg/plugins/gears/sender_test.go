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
