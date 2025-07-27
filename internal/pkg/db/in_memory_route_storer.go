// SPDX-FileCopyrightText: 2025 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package db

import (
	"context"
	"sync"
	"time"

	"github.com/xmidt-org/ears/internal/pkg/config"
	"github.com/xmidt-org/ears/internal/pkg/rtsemconv"
	"github.com/xmidt-org/ears/pkg/route"
	"github.com/xmidt-org/ears/pkg/tenant"
	"go.opentelemetry.io/otel"
)

type InMemoryRouteStorer struct {
	tenants map[string]map[string]*route.Config
	lock    *sync.RWMutex
}

func NewInMemoryRouteStorer(config config.Config) *InMemoryRouteStorer {
	return &InMemoryRouteStorer{
		tenants: make(map[string]map[string]*route.Config),
		lock:    &sync.RWMutex{},
	}
}

func (s *InMemoryRouteStorer) GetAllRoutes(ctx context.Context) ([]route.Config, error) {
	s.lock.RLock()
	defer s.lock.RUnlock()
	_, span := CreateSpan(ctx, "getRoutes", rtsemconv.DBSystemInMemory)
	defer span.End()
	span.SetAttributes(rtsemconv.DBSystemInMemory)
	routes := make([]route.Config, 0)
	for _, tenant := range s.tenants {
		for _, r := range tenant {
			routes = append(routes, *r)
		}
	}
	return routes, nil
}

func (s *InMemoryRouteStorer) GetRoute(ctx context.Context, tid tenant.Id, id string) (route.Config, error) {
	s.lock.RLock()
	defer s.lock.RUnlock()
	_, span := CreateSpan(ctx, "getRoute", rtsemconv.DBSystemInMemory)
	defer span.End()
	empty := route.Config{}
	t, ok := s.tenants[tid.Key()]
	if !ok {
		return empty, &route.RouteNotFoundError{TenantId: tid, RouteId: id}
	}
	r, ok := t[id]
	if !ok {
		return empty, &route.RouteNotFoundError{TenantId: tid, RouteId: id}
	}
	newCopy := *r
	return newCopy, nil
}

func (s *InMemoryRouteStorer) GetAllTenantRoutes(ctx context.Context, id tenant.Id) ([]route.Config, error) {
	s.lock.RLock()
	defer s.lock.RUnlock()
	_, span := CreateSpan(ctx, "getTenantRoutes", rtsemconv.DBSystemInMemory)
	defer span.End()
	span.SetAttributes(rtsemconv.DBSystemInMemory)
	routes := make([]route.Config, 0)
	t, ok := s.tenants[id.Key()]
	if !ok {
		return routes, nil
	}
	for _, r := range t {
		routes = append(routes, *r)
	}
	return routes, nil
}

func (s *InMemoryRouteStorer) setRoute(r route.Config) {
	r.Modified = time.Now().Unix()
	var tenant map[string]*route.Config
	if t, ok := s.tenants[r.TenantId.Key()]; !ok {
		tenant = make(map[string]*route.Config)
		s.tenants[r.TenantId.Key()] = tenant
	} else {
		tenant = t
	}
	if existing, ok := tenant[r.Id]; !ok {
		r.Created = r.Modified
	} else {
		r.Created = existing.Created
	}
	tenant[r.Id] = &r
}

func (s *InMemoryRouteStorer) SetRoute(ctx context.Context, r route.Config) error {
	s.lock.Lock()
	defer s.lock.Unlock()
	tracer := otel.Tracer(rtsemconv.EARSTracerName)
	_, span := tracer.Start(ctx, "storeRoute")
	defer span.End()
	span.SetAttributes(rtsemconv.DBSystemInMemory)
	defer span.End()
	s.setRoute(r)
	return nil
}

func (s *InMemoryRouteStorer) SetRoutes(ctx context.Context, routes []route.Config) error {
	s.lock.Lock()
	defer s.lock.Unlock()
	_, span := CreateSpan(ctx, "storeRoutes", rtsemconv.DBSystemInMemory)
	defer span.End()
	span.SetAttributes(rtsemconv.DBSystemInMemory)
	for _, r := range routes {
		s.setRoute(r)
	}
	return nil
}

func (s *InMemoryRouteStorer) DeleteRoute(ctx context.Context, tid tenant.Id, id string) error {
	s.lock.Lock()
	defer s.lock.Unlock()
	_, span := CreateSpan(ctx, "deleteRoute", rtsemconv.DBSystemInMemory)
	defer span.End()
	span.SetAttributes(rtsemconv.DBSystemInMemory)
	t, ok := s.tenants[tid.Key()]
	if !ok {
		return nil
	}
	delete(t, id)
	return nil
}

func (s *InMemoryRouteStorer) DeleteRoutes(ctx context.Context, tid tenant.Id, ids []string) error {
	s.lock.Lock()
	defer s.lock.Unlock()
	_, span := CreateSpan(ctx, "deleteRoutes", rtsemconv.DBSystemInMemory)
	defer span.End()
	span.SetAttributes(rtsemconv.DBSystemInMemory)
	t, ok := s.tenants[tid.Key()]
	if !ok {
		return nil
	}
	for _, id := range ids {
		delete(t, id)
	}
	return nil
}
