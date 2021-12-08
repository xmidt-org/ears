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

package db

import (
	"context"
	"github.com/xmidt-org/ears/internal/pkg/config"
	"github.com/xmidt-org/ears/internal/pkg/rtsemconv"
	"github.com/xmidt-org/ears/pkg/fragments"
	"github.com/xmidt-org/ears/pkg/route"
	"github.com/xmidt-org/ears/pkg/tenant"
	"go.opentelemetry.io/otel"
	"sync"
)

type InMemoryFragmentStorer struct {
	tenants map[string]map[string]*route.PluginConfig
	lock    *sync.RWMutex
}

func NewInMemoryFragmentStorer(config config.Config) *InMemoryFragmentStorer {
	return &InMemoryFragmentStorer{
		tenants: make(map[string]map[string]*route.PluginConfig),
		lock:    &sync.RWMutex{},
	}
}

func (s *InMemoryFragmentStorer) GetAllFragments(ctx context.Context) ([]route.PluginConfig, error) {
	s.lock.RLock()
	defer s.lock.RUnlock()
	_, span := CreateSpan(ctx, "getFragments", rtsemconv.DBSystemInMemory)
	defer span.End()
	span.SetAttributes(rtsemconv.DBSystemInMemory)
	fragments := make([]route.PluginConfig, 0)
	for _, tenant := range s.tenants {
		for _, f := range tenant {
			fragments = append(fragments, *f)
		}
	}
	return fragments, nil
}

func (s *InMemoryFragmentStorer) GetFragment(ctx context.Context, tid tenant.Id, id string) (route.PluginConfig, error) {
	s.lock.RLock()
	defer s.lock.RUnlock()
	_, span := CreateSpan(ctx, "getFragment", rtsemconv.DBSystemInMemory)
	defer span.End()
	empty := route.PluginConfig{}
	t, ok := s.tenants[tid.Key()]
	if !ok {
		return empty, &fragments.FragmentNotFoundError{TenantId: tid, FragmentName: id}
	}
	f, ok := t[id]
	if !ok {
		return empty, &fragments.FragmentNotFoundError{TenantId: tid, FragmentName: id}
	}
	newCopy := *f
	return newCopy, nil
}

func (s *InMemoryFragmentStorer) GetAllTenantFragments(ctx context.Context, tid tenant.Id) ([]route.PluginConfig, error) {
	s.lock.RLock()
	defer s.lock.RUnlock()
	_, span := CreateSpan(ctx, "getTenantFragments", rtsemconv.DBSystemInMemory)
	defer span.End()
	span.SetAttributes(rtsemconv.DBSystemInMemory)
	fragments := make([]route.PluginConfig, 0)
	t, ok := s.tenants[tid.Key()]
	if !ok {
		return fragments, nil
	}
	for _, r := range t {
		fragments = append(fragments, *r)
	}
	return fragments, nil
}

func (s *InMemoryFragmentStorer) setFragment(tid tenant.Id, f route.PluginConfig) {
	var tenant map[string]*route.PluginConfig
	if t, ok := s.tenants[tid.Key()]; !ok {
		tenant = make(map[string]*route.PluginConfig)
		s.tenants[tid.Key()] = tenant
	} else {
		tenant = t
	}
	tenant[f.FragmentName] = &f
}

func (s *InMemoryFragmentStorer) SetFragment(ctx context.Context, tid tenant.Id, r route.PluginConfig) error {
	s.lock.Lock()
	defer s.lock.Unlock()
	tracer := otel.Tracer(rtsemconv.EARSTracerName)
	_, span := tracer.Start(ctx, "storeFragment")
	defer span.End()
	span.SetAttributes(rtsemconv.DBSystemInMemory)
	defer span.End()
	s.setFragment(tid, r)
	return nil
}

func (s *InMemoryFragmentStorer) SetFragments(ctx context.Context, tid tenant.Id, fragments []route.PluginConfig) error {
	s.lock.Lock()
	defer s.lock.Unlock()
	_, span := CreateSpan(ctx, "storeFragments", rtsemconv.DBSystemInMemory)
	defer span.End()
	span.SetAttributes(rtsemconv.DBSystemInMemory)
	for _, f := range fragments {
		s.setFragment(tid, f)
	}
	return nil
}

func (s *InMemoryFragmentStorer) DeleteFragment(ctx context.Context, tid tenant.Id, id string) error {
	s.lock.Lock()
	defer s.lock.Unlock()
	_, span := CreateSpan(ctx, "deleteFragment", rtsemconv.DBSystemInMemory)
	defer span.End()
	span.SetAttributes(rtsemconv.DBSystemInMemory)
	t, ok := s.tenants[tid.Key()]
	if !ok {
		return nil
	}
	delete(t, id)
	return nil
}

func (s *InMemoryFragmentStorer) DeleteFragments(ctx context.Context, tid tenant.Id, ids []string) error {
	s.lock.Lock()
	defer s.lock.Unlock()
	_, span := CreateSpan(ctx, "deleteFragments", rtsemconv.DBSystemInMemory)
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
