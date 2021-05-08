package syncer

import (
	"context"
	"github.com/xmidt-org/ears/pkg/tenant"
)

const (
	EARS_ADD_ITEM_CMD       = "add"
	EARS_REMOVE_ITEM_CMD    = "rem"
	EARS_STOP_LISTENING_CMD = "stop"
)

const (
	ITEM_TYPE_ROUTE  = "route"
	ITEM_TYPE_TENANT = "tenant"
)

type SyncCommand struct {
	Cmd        string
	ItemType   string
	ItemId     string
	InstanceId string
	Sid        string
	Tenant     tenant.Id
}

type (
	LocalSyncer interface {
		// SyncRoute
		SyncItem(ctx context.Context, tid tenant.Id, itemId string, add bool) error
	}

	DeltaSyncer interface {
		// StartListeningForSyncRequests
		StartListeningForSyncRequests() // do we need this still?
		// StopListeningForSyncRequests
		StopListeningForSyncRequests() // do we need this still?
		// RegisterLocalTableSyncer
		RegisterLocalSyncer(itemType string, localTableSyncer LocalSyncer)
		// UnregisterLocalTableSyncer
		UnregisterLocalSyncer(itemType string, localTableSyncer LocalSyncer)
		// PublishSyncRequest
		PublishSyncRequest(ctx context.Context, tenantId tenant.Id, itemType string, itemId string, add bool)
		// GetInstanceCount
		GetInstanceCount(ctx context.Context) int
	}
)
