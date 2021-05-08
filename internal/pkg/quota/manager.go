package quota

import (
	"context"
	"github.com/xmidt-org/ears/internal/pkg/syncer"
	"github.com/xmidt-org/ears/pkg/ratelimit"
	"github.com/xmidt-org/ears/pkg/tenant"
	"sync"
)

type QuotaManager struct {
	limiters     map[string]ratelimit.RateLimiter
	tenantStorer tenant.TenantStorer
	syncer       syncer.DeltaSyncer
	lock         *sync.Mutex
}

func NewQuotaManager(tenantStorer tenant.TenantStorer, syncer syncer.DeltaSyncer) *QuotaManager {
	return &QuotaManager{
		limiters:     make(map[string]ratelimit.RateLimiter),
		tenantStorer: tenantStorer,
		syncer:       syncer,
		lock:         &sync.Mutex{},
	}
}

// Wait until rate limiter allows it to go through or context cancellation
func (m *QuotaManager) Wait(ctx context.Context, tid tenant.Id) error {
	//TODO
	return nil
}

func (m *QuotaManager) getLimiter(ctx context.Context, tid tenant.Id) (ratelimit.RateLimiter, error) {
	m.lock.Lock()
	defer m.lock.Unlock()

	limiter, ok := m.limiters[tid.Key()]
	if ok {
		return limiter, nil
	}
	//TODO implement
	return nil, nil
	//config, err := m.tenantStorer.GetConfig(ctx, tid)
}
