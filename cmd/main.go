package main

import (
	"fmt"

	"github.com/xmidt-org/ears/internal"
)

func main() {
	re := new(internal.RoutingEntry)
	fmt.Printf("hello world %s\n", re.PartnerId)
}
