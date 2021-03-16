# Routing Table Synchronization

## Architecture

There are four main concerns:

* startup
* shutdown
* delta synchronization
* periodic synchronization

![architecture](img/sync/architecture.png)

On startup, RoutingTableManager (RTM) will load, register and run all routes from the storage layer.
RTM manages several in-memory data structures to keep track of all known route configurations and 
their runtime information.

RTM registers itself as an observer of a DeltaPushSyncer (DTM). When another EARS instance modifies 
the shared routing table, the RTM receives a delta event from the DTM. A delta event contains a 
route ID and an operation code (add or remove) as payload. If a new route has been added, RTM will 
load the route from the storage layer by ID and register and run it. If an existing route has been
removed, RTM will stop and unregister the route.

Despite best efforts, over time the in-memory routing table managed by the RTM may become inconsistent
with the routing table persisted in the storage layer (source of truth). Therefore, RTM will check 
periodically, if there is any drift between the persisted version of the routing table and the 
RTM's in-memory version. If any inconsistencies are detected, RTM will repair those accordingly.

On shutdown, RTM will stop and unregister all routes.

## Delta Sync

Currently we have a Redis pub-sub based implementation of the delta syncer interface and also an
in-memory implementation to be used as a mock-implementation for unit tests. The documentation here
will focus on the Redis implementation.

![architecture](img/sync/syncdelta.png)

The Redis implementation utilizes two Redis topics, _ears_sync_ and _ears_ack_. All EARS instances 
in a cluster subscribe to the _ears_sync_ topic to learn about routing table changes performed by
other EARS instances. By contrast, if an ERAS instance is actively processing a CRUD operation, it 
will publish an event on the _ears_sync_ channel to notify all other instances.

Delta event consist of comma separated operation code, rules id, instance id and a unique session id.
The session id is necessary to allow for concurrent CRUD operation originating from the same EARS
instance.

1. An EARS Instance (IS1) receives an AddRoute() API call
1. IS1 stores the new route config in the shared persistence layer, here also Redis: `HSET routes rid config` 
1. IS1 publishes a delta event on _ears_sybc_: `PUBLISH ears_sync add,rid,iid,sid`
1. IS1 register and runs the route
1. Meanwhile, IS2 receives the delta event and makes sure the instance ID in the event is not idenitical 
   with hits own instance ID
1. IS2 then loads the route config from the persistence layer
1. IS2 now also register and runs the route
1. IS2 sends an ack message back through the _ears_ack_ channel using its own instance ID and teh same
   session ID: `PUBLISH ears_sync add,rid,iid,sid`
1. Once IS1 has collected acks from all other instances for the current session ID it can be sure the 
   new route information has been fully propagated
   
Note that the acking currently serves no real purpose, but we may choose to make use of it in the future.
Currently, the publishing of delta events follows a fire and forget strategy and relies on the periodic
syncing process for repairs of any inconsistencies. For that reason the operation of publishing and collecting 
acks does _not_ have to be implemented as a blocking operation which simplifies testing.

## Periodic Sync

The Routing Table Manager (RTM) will check periodically, if there is any drift between the persisted version 
of the routing table and the RTM's in-memory version. If any inconsistencies are detected, these need to 
be repaired.

![architecture](img/sync/syncperiodic.png)

1. Every five minutes, the RTM will load the entire routing table from the persistence layer
1. The RTM will then iterate over all the running routes it holds in memory
1. If the RTM finds any running routes that don't have a counterpart with the same ID in 
   the persistence layer, it will stop and unregister those routes
1. If the RTM finds any running routes that do have a counterpart with the same ID in
   the persistence layer but a different configuration hash, it will also stop and unregister those routes
1. RTM will then iterate over all the routes it loaded from the persistence layer earlier; if it finds any
   routes that are not present in the RTM in-memory table, then those missing routes will be registered
   and run

## Startup

On startup the Routing Table Manager (RTM) will load all route configurations from the persistence layer
and register and run those. This process initializes all routes in the RTM's in memory routing table.

![architecture](img/sync/startup.png)

## Shhutdonw

On shutdown the Routing Table Manager (RTM) will iterate over all the live routes in its in-memory routing
table and will stop and unregister all of them.

![architecture](img/sync/teardown.png)

## Data Structures

The RouteConfig combines a route ID, a route configuration and a route hash. The hash is an MD5 hash over
the entire configuration excluding the route ID. Note that the route ID can be chosen by the caller of the
AddRoute() API. If the caller omits the route ID the RTM will choose the route hash as the route ID. This means
a route ID can be identical with the route hash but it doesn't have to be.

![architecture](img/sync/routeconfig.png)



![architecture](img/sync/liveroutewrapper.png)

![architecture](img/sync/table1.png)

![architecture](img/sync/table2.png)

![architecture](img/sync/table4.png)

![architecture](img/sync/table3.png)

![architecture](img/sync/table5.png)

## Adding Routes, Updating Routes, Deleting Routes

