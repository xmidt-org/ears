# Notes On Ack, Nack and Retries

## The Current Design And Its Implications

In a scenario where multiple routes share the same receiver, EARS follows a
different algorithm for ack (success case) and nack (failure case). 

After a 
single event fans out into multiple filter chains, EARS collects the acks
for the event. Only after all routes attached to a single receiver instance 
have concluded with an ack, the ack routine of the receiver gets executed.
This routine typically writes some metrics and cancels the event context.

However, if any one of the filters or senders in the route tree finishes with a
nack, the nack routine of the receiver gets executed immediately regardless of 
whether other legs of the route tree are still processing (fail fast). This means, 
depending on timing, that other routes may get canceled before they are able to 
deliver the event to their destination and ack the event. The net affect is that one route can 
"steal" events from other routes. This side effect only occurs within the same
application but not across applications and organizations. Also notice that this 
is only an issue when multiple routes share the same receiver and some routes end
in ack and some in nack.

The receiver orchestrates the lifecycle of an event. In case of all acks, 
the receiver is clearly done with the event and can safely remove it from 
the event source (what exactly that means is protocol specific).
However, the main purpose of a nack is to signal to the receiver
that event processing has failed and that a retry may have a chance of success. 
It is then up to the receiver to decide if a retry should actually be attempted or not.
This decision depends on the retry policy and the receiver configuration.
For example a receiver may implement a policy that allows up to three retries 
but no more than that. The motivation behind the fail-fast on nack is to reduce
the probability of duplicate event delivery. This is because a single nack
will trigger a retry impacting all routes sharing the same receiver. Therefore
the event will be redelivered even to routes that would have concluded or actually 
did conclude in an ack.

## Beware Of Filter Nacks

If a filter decides to consume an event (due to filtering) it should ack the event.
If a filter decides to pass an event on it should neither ack nor nack the event.
Rarely do we have a scenario though where it makes sense for a filter to nack
the event - because that would signal to the receiver that there is hope for success
when retrying the event. However most filter logic is completely deterministic and
throwing the same event at the same filter will lead to the same outcome. A 
notable exception is the webservice filter which makes an http request to an external
service. If reaching an external service fails, a retry may be justified, therefore
the filter may choose to nack.

## Sender Retries - A Strategy For Avoiding Duplicates

One idea to avoid unwanted redelivery (duplicates), is to implement an optional retry policy 
at the sender level. For example, if a sender fails to deliver an event to its destination,
it could quietly perform up to N retries (meaning without ack or nack). If any of the 
retries succeed we can happily ack this back to the receiver. Even if the Nth retry 
attemp fails we should still ack the event because further retries have diminished hope
for success. By delegating retries to the sender we avoid including other routes in the 
retry attempts. The same argument could be made for the webservice filter. By executing
retries locally we can keep the receiver in the dark about what is going on and avoid 
potential duplicates.

## What About Visibility?

## Virtual Receivers, SenderReceivers

## Durable Events Using External Queues

