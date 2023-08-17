// Copyright 2021 Comcast Cable Communications Management, LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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

type EarsMetric struct {
	SuccessCount    int
	ErrorCount      int
	FilterCount     int
	SuccessVelocity int
	ErrorVelocity   int
	FilterVelocity  int
	LastEventTs     int64
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
		//
		WriteMetrics(id string, metric *EarsMetric)
		//
		ReadMetrics(id string) *EarsMetric
	}
)
