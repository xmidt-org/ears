# Routes

Routes are a key concept of EARS. As a user of EARS your business is adding new routes, updating existing 
routes and deleting routes. To make the most of EARS, it is crucial to gain an in-depth understanding of how 
routes work in EARS.

Routes connect data sources to data destinations, for example you can connect an SQS queue to a Kafka topic. 
Events travelling through such a route can optionally be transformed or filtered according to the configuration details of 
the route.
 
A route consists of a linear sequence of a receiver plugin, followed by zero, one or more filter plugins (the filter chain) 
and a sender plugin. Note that the filter chain can be empty in which case the route simply consists of a receiver 
directly connected to a sender. In either case a route processes events asynchronously, one at a time in sequence.

![image route](img/route_2.png)

A real life example may look like this:

![image route](img/route.png)

The receiver plugin of a route receives events from an event source. The receiver plugin encapsulates all protocol 
specific implementation and configuration details. EARS currently offers receiver plugins for Kafka, Kinesis, SQS, Redis
Pub/Sub and HTTP. The receiver plugin passes each event it receives from its event source to the filter chain.
Each filter plugin in the filter chain is configured to perform a specific task. EARS offers a small library of highly 
generic and configurable filters to perform various types of payload transformations and filtering. For example, a match
filter will drop events that do 
not meet predefined criteria such as presence or absence of a certain key value pair in the payload etc. After
traveling through the filter chain, events finally arrive at the sender plugin which will then deliver the events to
a data destination. Again EARS offers a set of configurable sender plugins for Kafka, Kinesis, SQS, Redis
Pub/Sub and HTTP that encapsulate all destination protocol specific implementation details.

Conceptually you may think of a route as a linear flow where a receiver plugin is followed by some filter plugins 
which are followed by a sender plugin. There are no forks or loops allowed in the structure of the route. 

## JSON or YAML

When manipulating routes using the EARS API you may submit route configurations either using JSON or YAML
encoding. Often JSON route configurations suffice but whenever a route contains multi-line strings such as 
lengthy JavaScript in a _js_ filter then using YAML encoding may result in more readable route configurations.

## Routing Table Synchronization

To scale horizontally, EARS stores all routes in a central shared routing table which is treated as the source 
of truth by all EARS instances. Every module in EARS is defined as a golang interface and the persistence layer 
for the routing table is no exception. By default, EARS comes with an in memory implementation of the routing table. 
This is useful for single instance deployments and for debugging. For production use EARS offers a DynamoDB 
implementation of the same interface. In essence the persistence layer for the routing table is fully pluggable 
and you can provide your own implementation to support other types of storage layer.

![image route](img/routing_table_sync.png)

When an EARS instance is started, it will load the entire routing table from storage (usually DynamoDB) and 
keep a local copy of it in 
an in-memory cache. When adding or updating a route, the AddRoute() API will be executed on one of the 
EARS instances behind a load balancer. This EARS instance will then orchestrate the update of the routing table
across the entire EARS cluster. First it updates the routing table in DynamoDB. It then sends an update notification
to all other EARS instances which will update their local copies of the routing table. This way all EARS instances
should be in sync with the latest version of the routing table very rapidly. In addition, each EARS instance will
periodically scan its routing table for differences with the reference routing table in DynamoDB by comparing 
route hashes. If an EARS instance detects any inconsistencies it will repair its local copy by updating bad routes
with the correct configurations from DynamoDB.

## Stream Sharing

Imagine you have two different routes that read from the same data source, for example an SQS queue, using the exact
same parameters in their receiver configuration.  

![image route](img/stream_sharing_1.png)

In a situation like this, EARS will detect this an only create one instance of the SQS receiver plugin which will 
then be shared by all routes which have identical receiver configurations. In this example, every event received
from SQS will be cloned by the SQS receiver and each filter chain will receive its own copy of the event. 
Effectively, the SQS receiver will fan out each incoming event to all routes that have expressed an interest.
When just working with two routes, stream sharing may not be important. But once you work with thousands of routes
all reading from the same high-throughput topic, stream sharing becomes essential for performance reasons.

Note, that stream sharing will only occur within the same org ID and app ID. There is no stream sharing across
tenant boundaries.

![image route](img/stream_sharing_2.png)

Also note, that stream sharing is done by EARS automatically in the background. As a user of EARS you do not have to 
do anything for this to happen and when creating or removing routes you still treat every route individually as if
no sharing occurred. Internally EARS will maintain a reference counter to keep track of how many routes are
receiving events from a particular receiver instance, and whenever that reference count drops to zero, due to deletion
of routes, EARS will automatically shut down a receiver plugin that is not needed any loger.

In some situations, however, you may not want string sharing. You can force EARS to not do stream sharing by
choosing unique receiver plugin names. Each receiver plugin configuration has an optional name field. By choosing
different names for the receiver plugins of different routes, the receiver plugin hashes will be different
and therefore, stream sharing will not be applied even if the receiver configurations of two routes are otherwise
identical.







