package app

import "context"

type (
	// A RoutingEntry represents an entry in the EARS routing table
	RoutingTableEntry struct {
	}
)

type (

	// A RoutingTableManager supports modifying and querying an EARS routing table
	RoutingTableManager interface {
		AddRoute(ctx context.Context, entry *RoutingTableEntry) error // idempotent operation to add a routing entry to a local routing table 		// get hash for local version of routing table
	}
)
