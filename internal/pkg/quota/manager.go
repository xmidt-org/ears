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

package quota

import (
	"context"
	"errors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/xmidt-org/ears/internal/pkg/config"
	"github.com/xmidt-org/ears/internal/pkg/syncer"
	"github.com/xmidt-org/ears/pkg/tenant"
	"sync"
	"time"
)

type QuotaManager struct {
	limiters           map[string]*QuotaLimiter
	tenantStorer       tenant.TenantStorer
	syncer             syncer.DeltaSyncer
	lock               *sync.Mutex
	backendLimiterType string
	redisAddr          string
	logger             *zerolog.Logger

	ticker *time.Ticker
	done   context.CancelFunc
	ctx    context.Context
}

const LimiterTypeRedis = "redis"
const LimiterTypeInMemory = "inmemory"

func NewQuotaManager(logger *zerolog.Logger, tenantStorer tenant.TenantStorer, syncer syncer.DeltaSyncer, config config.Config) (*QuotaManager, error) {
	backendLimiterType := config.GetString("ears.ratelimiter.type")

	if backendLimiterType != LimiterTypeRedis && backendLimiterType != LimiterTypeInMemory {
		return nil, &BadConfigError{"ears.ratelimiter.type", backendLimiterType}
	}

	redisAddr := ""
	if backendLimiterType == LimiterTypeRedis {
		redisAddr = config.GetString("ears.ratelimiter.endpoint")
		if redisAddr == "" {
			return nil, &ConfigNotFoundError{"ears.ratelimiter.endpoint"}
		}
	}

	return &QuotaManager{
		limiters:           make(map[string]*QuotaLimiter),
		tenantStorer:       tenantStorer,
		syncer:             syncer,
		lock:               &sync.Mutex{},
		backendLimiterType: backendLimiterType,
		redisAddr:          redisAddr,
		logger:             logger,
	}, nil
}

func (m *QuotaManager) Start() {
	m.syncer.RegisterLocalSyncer("tenant", m)

	//Start backup quota syncer that wakes up every minute to sync on tenant quota
	ticker := time.NewTicker(time.Minute)
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)

	m.done = cancel
	m.ctx = ctx
	m.ticker = ticker

	go func() {
		for {
			m.logger.Info().Str("op", "PeriodicQuotaSync").Msg("Periodically sync tenant quotas")
			select {
			case <-m.ctx.Done():
				return
			case <-m.ticker.C:
				m.syncAllItems()
			}
		}
	}()
}

func (m *QuotaManager) Stop() {
	m.syncer.UnregisterLocalSyncer("tenant", m)

	if m.ticker != nil {
		m.ticker.Stop()
		m.done()
	}
}

// Wait until rate limiter allows it to go through or context cancellation
func (m *QuotaManager) Wait(ctx context.Context, tid tenant.Id) error {
	limiter, err := m.getLimiter(ctx, tid)
	if err != nil {
		return nil
	}
	return limiter.Wait(ctx)
}

func (m *QuotaManager) TenantLimit(ctx context.Context, tid tenant.Id) int {
	limiter, err := m.getLimiter(ctx, tid)
	if err != nil {
		return 0
	}
	return limiter.Limit()
}

func (m *QuotaManager) SyncItem(ctx context.Context, tid tenant.Id, itemId string, add bool) error {
	limiter, err := m.getLimiter(ctx, tid)
	if err != nil {
		return err
	}
	config, err := m.tenantStorer.GetConfig(ctx, tid)
	tenantRqs := 0
	if err != nil {
		var tenantNotFound *tenant.TenantNotFoundError
		if !errors.As(err, &tenantNotFound) {
			return err
		}
	} else {
		tenantRqs = config.Quota.EventsPerSec
	}
	return limiter.SetLimit(tenantRqs)
}

//PublishQuota publishes tenant quota to ratelimiters in all nodes so they can sync to the new quota
func (m *QuotaManager) PublishQuota(ctx context.Context, tid tenant.Id) error {
	err := m.SyncItem(ctx, tid, "ignored", true)
	if err != nil {
		log.Ctx(ctx).Error().Str("op", "PublishQuota").Str("action", "SyncItem").Str("error", err.Error()).Msg("Error syncing local quota")
	}

	m.syncer.PublishSyncRequest(ctx, tid, syncer.ITEM_TYPE_TENANT, "ignored", true)
	return nil
}

func (m *QuotaManager) getLimiter(ctx context.Context, tid tenant.Id) (*QuotaLimiter, error) {
	m.lock.Lock()
	defer m.lock.Unlock()

	limiter, ok := m.limiters[tid.Key()]
	if ok {
		return limiter, nil
	}

	config, err := m.tenantStorer.GetConfig(ctx, tid)
	tenantRqs := 0
	if err != nil {
		var tenantNotFound *tenant.TenantNotFoundError
		if !errors.As(err, &tenantNotFound) {
			return nil, err
		}
	} else {
		tenantRqs = config.Quota.EventsPerSec
	}

	instanceCount := m.syncer.GetInstanceCount(ctx)
	if instanceCount == 0 {
		return nil, &NoEarsInstances{}
	}

	initialRqs := tenantRqs / instanceCount

	limiter = NewQuotaLimiter(tid, m.backendLimiterType, m.redisAddr, initialRqs, tenantRqs)
	m.limiters[tid.Key()] = limiter
	return limiter, nil
}

func (m *QuotaManager) syncAllItems() {
	//make a copy of quota limiters
	m.lock.Lock()
	limiters := make([]*QuotaLimiter, len(m.limiters))
	i := 0
	for _, limiter := range m.limiters {
		limiters[i] = limiter
		i++
	}
	m.lock.Unlock()

	for _, limiter := range limiters {
		m.SyncItem(m.ctx, limiter.tid, "ignored", true)
	}
}
