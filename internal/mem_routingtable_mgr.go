/**
 *  Copyright (c) 2020  Comcast Cable Communications Management, LLC
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package internal

import (
	"context"
	"errors"
	"sync"
)

type (
	InMemoryRoutingTableManager struct {
		//TODO: add index by source plugin
		routingTableIndex RoutingTableIndex
		lock              sync.RWMutex
	}
)

// NewInMemoryRoutingTableManager creates a new local in memory routing table cache
func NewInMemoryRoutingTableManager() *InMemoryRoutingTableManager {
	mgr := new(InMemoryRoutingTableManager)
	mgr.routingTableIndex = make(map[string]*RoutingTableEntry)
	mgr.lock = sync.RWMutex{}
	return mgr
}

// AddRoute adds a route to the routing table
func (mgr *InMemoryRoutingTableManager) AddRoute(ctx context.Context, entry *RoutingTableEntry) error {
	mgr.lock.Lock()
	defer mgr.lock.Unlock()
	if entry == nil {
		return errors.New("missing routing table entry")
	}
	if err := entry.Validate(ctx); err != nil {
		return err
	}
	if err := entry.Initialize(ctx); err != nil {
		return err
	}
	entry.tblMgr = mgr
	mgr.routingTableIndex[entry.Hash(ctx)] = entry
	return nil
}

// RemoveRoute removes a route from the routing table
func (mgr *InMemoryRoutingTableManager) RemoveRoute(ctx context.Context, entry *RoutingTableEntry) error {
	mgr.lock.Lock()
	defer mgr.lock.Unlock()
	if entry == nil {
		return errors.New("missing routing table entry")
	}
	if err := entry.Validate(ctx); err != nil {
		return err
	}
	delete(mgr.routingTableIndex, entry.Hash(ctx))
	return nil
}

// ReplaceAllRoutes replaces all routes
func (mgr *InMemoryRoutingTableManager) ReplaceAllRoutes(ctx context.Context, entries []*RoutingTableEntry) error {
	m := make(map[string]*RoutingTableEntry)
	for _, entry := range entries {
		if err := entry.Validate(ctx); err != nil {
			return err
		}
		m[entry.Hash(ctx)] = entry
	}
	mgr.routingTableIndex = m
	return nil
}

// Validate validates all entries in the routing table
func (mgr *InMemoryRoutingTableManager) Validate(ctx context.Context) error {
	for _, entry := range mgr.routingTableIndex {
		if err := entry.Validate(ctx); err != nil {
			return err
		}
	}
	return nil
}

// Hash calculates hash over all entries in the routing table
func (mgr *InMemoryRoutingTableManager) Hash(ctx context.Context) string {
	hash := ""
	for _, entry := range mgr.routingTableIndex {
		hash = hash + entry.Hash(ctx)
	}
	return hash
}

// GetAll Routes returns complete routing table
func (mgr *InMemoryRoutingTableManager) GetAllRoutes(ctx context.Context) ([]*RoutingTableEntry, error) {
	mgr.lock.RLock()
	defer mgr.lock.RUnlock()
	tbl := make([]*RoutingTableEntry, len(mgr.routingTableIndex))
	idx := 0
	for _, entry := range mgr.routingTableIndex {
		tbl[idx] = entry
	}
	return tbl, nil
}

// GetRoutesBySourcePlugin gets all routes used by a given input plugin
func (mgr *InMemoryRoutingTableManager) GetRoutesBySourcePlugin(ctx context.Context, plugin *InputPlugin) ([]*RoutingTableEntry, error) {
	mgr.lock.RLock()
	defer mgr.lock.RUnlock()
	tbl := make([]*RoutingTableEntry, len(mgr.routingTableIndex))
	for _, entry := range mgr.routingTableIndex {
		if entry.Source.Hash(ctx) == plugin.Hash(ctx) {
			tbl = append(tbl, entry)
		}
	}
	return tbl, nil
}

// GetRoutesByDestinationPlugin gets all routes used by a given output plugin
func (mgr *InMemoryRoutingTableManager) GetRoutesByDestinationPlugin(ctx context.Context, plugin *OutputPlugin) ([]*RoutingTableEntry, error) {
	mgr.lock.RLock()
	defer mgr.lock.RUnlock()
	tbl := make([]*RoutingTableEntry, len(mgr.routingTableIndex))
	for _, entry := range mgr.routingTableIndex {
		if entry.Destination.Hash(ctx) == plugin.Hash(ctx) {
			tbl = append(tbl, entry)
		}
	}
	return tbl, nil
}

// GetRoutesForEvent gets all routes for a given event
func (mgr *InMemoryRoutingTableManager) GetRoutesForEvent(ctx context.Context, event *Event) ([]*RoutingTableEntry, error) {
	return mgr.GetRoutesBySourcePlugin(ctx, event.Source)
}

// GetRouteCount gets size of routing table
func (mgr *InMemoryRoutingTableManager) GetRouteCount(ctx context.Context) int {
	return len(mgr.routingTableIndex)
}
