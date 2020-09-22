package main

import (
	"fmt"

	"github.com/xmidt-org/ears/internal"
)

func main() {
	re := new(internal.RoutingTableEntry)
	fmt.Printf("hello world %s\n", re.OrgId)
}
