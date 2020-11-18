# Receiver Fanout

#### The Goals (are there more?)
1. Under <strong>normal</strong> operation, slower processing routes should not affect the speed of fast processing route
2. Under <strong>normal</strong> operation, a receiver should be able to ingest events at optimal rate
3. In the cases where one or more routes are congested or backed up (abnormal), it is OK to have degraded performance as long as:
    * There are alerts to notify Support/DevOps
    * EARs does not blow up (panic, OOM, etc)
    * The performance impact is only limited to the particular receiver

## Proposal 1
Code snippet in question (none working function for demo only)
```go
func (a *PerReceiver) next(ctx context.Context, event Event) error {
    wg := &sync.WaitGroup{}
    for _, r := a.routes {
        wg.Add(1)
        go func() {
            r.Process(ctx, event)
            wg.Done()
        }
    }
    wg.Wait()
}
```

The idea with the above code is that all routes are processed in parallel and that they are not subject to the processing time of the other routes. 

However, it does not meet goal number 2. Because the next function only returns when every route finishes processing (whatever that means), the latency of the next function is now bounded by the slowest processing route. Assuming that the receiver calls next sequentially for every event, the  receiver's ingestion rate is also bounded by the latency of the next function.

#### Observation
The next function should return as fast as possible. Its latency should be such that it is as fast or faster than the event ingestion rate. For example, if a receiver's event ingestion rate is 1000 events/sec, then the latency of the next function should be no greater than 1ms.

## Proposal 2
What if we remove the WaitGroup? The next function will now return very quickly.
```go
func (a *PerReceiver) next(ctx context.Context, event Event) error {
    for _, r := a.routes {
        go func() {
            r.Process(ctx, event)
        }
    }
}
```
However, it does not meet goal number 3. If a route becomes congested (r.Process takes a long time to return), it will result in goroutine leaks. At some point, we may run out of memory and EARs will crash.

#### Observation
Unbounded goroutine invocation (ones without coordination construct like WaitGroup) will impact system stability.

## Proposal 3
What if we cap the number of goroutines per receiver (or per route or per tenant)?

A per receiver implementation:
```go
func (a *PerReceiver) next(ctx context.Context, event Event) error {
    for _, r := a.routes {
        //the function block if max number of goroutine is reached per receiver
        a.countCond.l.Lock() //a.countCond is *sync.Cond
        for a.count >= MAX_COUNT {
            a.countCond.Wait()
        }
        a.count++
        go func() {
            r.Process(ctx, event)

            //decrement the goroutine count
            a.countCond.l.Lock()
            a.count--
            a.countCond.Signal()
            a.countCond.l.Unlock()
        } 
        a.countCond.l.UnLock()
    }
}
```

#### Observation
* Should be able to meet all of our goals
* Goroutines per route: This will prevent abnormal route affecting normal route with in the same receiver to some extent. At some point, if the back-pressure is too much, it may still affect all routes of the same receiver source. It will not affect routes with different receiver source. 
* Goroutines shared between routes in same receiver source: Abnormal route will saturate all goroutines and affect normal routes of the same reciever source. It will not affect routes with different receiver source.
* Goroutines shared between routes in same tenant: Abnormal route will saturate all goroutines and affect all routes within the tenant

## Proposal 4
What about channels and worker pools? (at receiver, route, or tenant level).

The implementation below has a dedicated buffer channel and worker pool at receiver level:
```go
type Work struct {
    r     Route
    event Event   
}
func (a *PerReceiver) next(ctx context.Context, event Event) error {
    for _, r := routes {
        work := &Work{
            r:     r,
            event: event,
        }
        a.ch <- work  //a.ch is a buffered channel
    }
}

func (a *PerReceiver) worker() {
    for work := range a.ch {
        work.r.Process(work.r.event)
    }
} 
```

#### Observation
* Should be able to meet all our goals
* Trade-off between proposal 3 and 4 are:
    * pre-allocated goroutine vs on-demand goroutine
    * buffered channel may absorb temporary down stream congestion without impacting receiver ingestion rate