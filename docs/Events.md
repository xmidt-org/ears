# EARS Event LifeCycle

Current Event interface:
```go
type Event interface {
    //Get the event payload
    Payload() interface{}
    
    //Set the event payload
    //Will return an error if the event is done
    SetPayload(payload interface{}) error
    
    //Acknowledge that the event is done and further actions on
    //the event are no longer possible.
    //Proposal 2: We do not expose Ack to the plugin developers.
    //            Instead, we do the ack for them (see below)
    Ack()
    
    //Duplicate the event. (Do we deep-copy payload?)
    //Dup fails and return an error if the event is done
    Dup() (Event, error)
}

type NewEventerer interface {
    NewEvent(payload interface{}) (Event, error)
}
```

In EARS, an event originated from a receiver. Lets call it a <strong>root event</strong>. As a root event journeys through pipelines from a receiver to senders, it can manifest itself into many events. Lets call these <strong>child events</strong>. 

Conceptually, a root event and its child events form an event tree.

Let's define the lifecycle of an event in an event tree:

An event is created from:
* a receiver as a root event
* a parent event as a child event

An event is terminated when:
* a sender sends it out
* a filterer filters it out

As an event travels, it can:
1. stay unchanged
2. be filtered
3. update its payload
4. update its context
5. produce child events.

Optionally, an event can be <strong>acknowledged</strong>. When all events in an event tree are acknowledged, the root event is considered <strong>handled</strong>, and the corresponding receiver that generates this event can safely clean it up (for example, from an upstream queue) where applicable.

## Receiver (maybe external developers)

A receiver creates a root event from a payload and a context:
```go
NewEvent(ctx context.Context, payload interface{}) (Event, error)
```

A root event created must have an associated context. A receiver is responsible for creating the context and for adding necessary context information (timeout, tracing, etc). A receiver will not be notified when the root event is handled. When a root event is instantiated this way, it means that the receiver is setup for fire and forget, so it does not care if an event is handled or not

A receiver creates a root event with acknowledgement:
```go
NewEventWithAcknowledgement(ctx context.Context, payload interface{}, handledFn func(), errFn func(error)) (Event, error)
```

Here is a possible implementation demonstrating on how this can work:
```go
  for {
  	//Receive payload from upstream sources (kafka, SQS, etc)
    payload := r.recvPayload()
    
    //Generate payload context with logging/tracing/metric information
    ctx := r.newContext(payload)
    
    //Create a root event
    e := event.NewEventWithAcknowldegement(ctx, payload, 
    	func(){
    		//Success! The root event handled
       	    r.deletePayload(payload)
        }, 
        func(err error) {
        	//Fail to handle root event
    	    log.Ctx(ctx).Error().Str("error", err.Error()).Msg("Fail to handle event")
        })
    
    //Send the event downstream
    r.next(e)
  }
```

## Routing Manager Fanout to Routes (internal)

If there are multiple routes configured for a receiver, the routing manager will *fantout* the events to the routes. The routing manager should do this by deriving child events from the root event, one per route. Each child event should have its own context. Also, a child event should have a deep-copy of the payload to prevent routing cross-talk.

Currently, the event interface have a `Dup` function for creating a child event:
```go
    //Duplicate the event. (Do we deep-copy payload?)
    //Dup fails and return an error if the event is done
    Dup() (Event, error)
```

I am proposing that we rename the function to `Clone` and it takes in a context as its parameter:
```go
    //Create a child the event with payload deep-copied
    //Clone fails and return an error if the event is already acknowledged
    Clone(ctx context.Context) (Event, error)
```

An example on how routing manager can do the fanout:
```go
    for _, n := range nextFns {
        wg.Add(1)
        go func(fn pkgreceiver.NextFn) {
            ctx := e.Context()
        	
            //derive a new context for each route (adding tracing info)
            ctx = tracingInfo(ctx)

            //create a new child event
            childEvent, err := e.Clone(ctx)
            if err != nil {
              	errCh <- err
              	return
            }
            
            err := fn(childEvent)
            if err != nil {
                errCh <- err
            }
        }(n)
    }
```

Please note that `Clone` can be expensive as it needs to deep-copy the payload using reflections (is that true?). For optimization, we can imagine that if a receiver only have one route (probably the majority case), we can have a shallow-copy version of the `Clone` function:
```go
    //Create a child the event with payload shallow-copied
    //WithContext fails and return an error if the event is already acknowledged
    WithContext(ctx context.Context) (Event, error)
```

## Filters (maybe external developers)
Filter can affect events in multiple ways.

Case 1: Event unchanged (Pass through):
```go
func (f *Filter) Filter(evt event.Event) ([]event.Event, error) {
	return []event.Event{evt}, nil
}
```

Case 2: Event filtered
```go
func (f *Filter) Filter(evt event.Event) ([]event.Event, error) {
    return []event.Event{}, nil
}
```

Case 3: Updating event with new payload
```go
func (f *Filter) Filter(evt event.Event) ([]event.Event, error) {
	newPayload := process(evt)
	evt.SetPayload(newPayload)
    return []event.Event{evt}, nil
}
```

Case 4: Updating event with new context
```go
func (f *Filter) Filter(evt event.Event) ([]event.Event, error) {
	newCtx := updateTraceInfo(evt)
	newEvt, err := evt.WithContext(newCtx)
	if err != nil {
	    //error handling	
		return nil, err
    }
    return []event.Event{newEvt}, nil
}
```

Case 5: Event splitting
```go
func (f *Filter) Filter(evt event.Event) ([]event.Event, error) {
    newPayloads := process(evt)
    events := make([]event.Event, 0)
    for _, payload := range newPayloads {
    	ctx := updateTraceInfo(evt.Context(), payload)
    	
    	//notice that we are using WithContext here
    	childEvt, err := evt.WithContext(ctx)
    	if err != nil {
    	    //error handling
    		return nil, err
        }
        childEvt.SetPayload(payload)
    	
        events = append(events, childEvt)	
    }
    
    return events, nil
}
```

## Sender (maybe external developers)
Senders just need to send the event to downstream. It should not need to update the events (maybe?)

## Routing Manager Event Acknowledgement (internal developer)
An event as an `Ack` function to acknowledge that an event is done and should not be used further. (Think of it as a destructor for an event). Once an event is <strong>Acked</strong>, subsequent `SetPayload`, `Clone`, and `WithContext` call will return an error.

Event acknowledgements are necessary so that we know can track if all events in an event tree are acknowledged so that the root event is considered handled.

The question is then, who is going to call the `Ack` function?

There are two possible answers:
1. Whoever (receiver, filter, sender, route manager) handling the event must call `evt.Ack()` when it no longer wants to pass that event downstream. Or in sender's case, the event is sent.
2. Between the plugin boundaries, the routing manager implicitly figure out if events are still needed. For all the child events that are no longer need to be passed down the pipeline, routing manager calls the `evt.Ack()` for them.

## Updated Event Interface

```go
type Event interface {
    //Get the event payload
    Payload() interface{}
    
    //Set the event payload
    //Will return an error if the event is done
    SetPayload(payload interface{}) error
    
    //Acknowledge that the event is done and further actions on
    //the event are no longer possible.
    //Proposal 2: We do not expose Ack to the plugin developers.
    //            Instead, we do the ack for them (see below)
    Ack()

    //Create a child the event with payload deep-copied
    //Clone fails and return an error if the event is already acknowledged
    Clone(ctx context.Context) (Event, error)

    //Create a child the event with payload shallow-copied
    //WithContext fails and return an error if the event is already acknowledged
    WithContext(ctx context.Context) (Event, error)
}

type NewEventerer interface {
    NewEvent(ctx context.Context, payload interface{}) (Event, error)
    NewEventWithAcknowledgement(ctx context.Context, payload interface{}, handledFn func(), errFn func(error)) (Event, error)
}
```

## Problem with Event as an interface
