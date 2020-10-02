package main

import (
	"context"
	"fmt"

	"github.com/xmidt-org/ears/internal"
)

func main() {
	ctx := context.Background()
	var rtmgr internal.RoutingTableManager
	rtmgr = internal.NewInMemoryRoutingTableManager()
	fmt.Printf("ears has %d routes\n", rtmgr.GetRouteCount(ctx))
}
