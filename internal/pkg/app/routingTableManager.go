package app

import "context"

type DefaultRoutingTableManager struct {
}

func NewRoutingTableManager() RoutingTableManager {
	return &DefaultRoutingTableManager{}
}

func (r *DefaultRoutingTableManager) AddRoute(ctx context.Context, entry *RoutingTableEntry) error {
	//TODO implement me
	return nil
}
