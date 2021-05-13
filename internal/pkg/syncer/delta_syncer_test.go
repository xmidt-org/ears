package syncer_test

import (
	"context"
	"github.com/xmidt-org/ears/internal/pkg/syncer"
	"github.com/xmidt-org/ears/pkg/tenant"
	"sync"
	"testing"
	"time"
)

type LocalSyncCollector struct {
	sync.Mutex
	count int
}

func (l *LocalSyncCollector) SyncItem(ctx context.Context, tid tenant.Id, itemId string, add bool) error {
	l.Lock()
	defer l.Unlock()

	l.count++
	return nil
}

func (l *LocalSyncCollector) Count() int {
	l.Lock()
	defer l.Unlock()

	return l.count
}

func (l *LocalSyncCollector) Reset() {
	l.Lock()
	defer l.Unlock()

	l.count = 0
}

func (l *LocalSyncCollector) ValidateCount(expectedCount int, t *testing.T) {
	time.Sleep(time.Millisecond * 100)

	//Make sure we collected at least 4 SyncItem events or if 500 ms has elapsed
	numTries := 0
	for l.Count() < expectedCount && numTries < 5 {
		time.Sleep(time.Millisecond * 100)
		numTries++
	}
	if l.Count() != expectedCount {
		t.Fatalf("Expect collect 4 SyncRequest, but get %d instead\n", l.Count())
	}
}

func testSyncers(newSyncer func() syncer.DeltaSyncer, t *testing.T) {

	collector := LocalSyncCollector{}
	ctx := context.Background()
	tid := tenant.Id{"myOrg", "myApp"}

	//Setup 5 syncer
	syncers := make([]syncer.DeltaSyncer, 5)
	for i := 0; i < 5; i++ {
		syncers[i] = newSyncer()
		syncers[i].StartListeningForSyncRequests()
		syncers[i].RegisterLocalSyncer("test", &collector)
	}

	instanceCount := syncers[0].GetInstanceCount(ctx)
	if instanceCount != 5 {
		t.Fatalf("Expect 5 instance but get %d instead\n", instanceCount)
	}

	//Case 1: publish sync event
	syncers[3].PublishSyncRequest(ctx, tid, "test", "testId", true)
	collector.ValidateCount(4, t)

	//Case 2: publish to a different itemType
	collector.Reset()
	syncers[1].PublishSyncRequest(ctx, tid, "nowhere", "testId", true)
	collector.ValidateCount(0, t)

	//Case 3: validate that a publisher does not get its own sync event
	collector.Reset()
	publisherCollector := LocalSyncCollector{}
	syncers[2].UnregisterLocalSyncer("test", &collector)
	syncers[2].RegisterLocalSyncer("test", &publisherCollector)
	syncers[2].PublishSyncRequest(ctx, tid, "test", "testId", false)
	collector.ValidateCount(4, t)
	publisherCollector.ValidateCount(0, t)

	//Teardown
	for _, syncer := range syncers {
		syncer.StopListeningForSyncRequests()
	}
}
