# Routes

# Routes

Routes are the core concept of EARS. As a user of EARS your business is adding new routes, updating existing 
routes and deleting routes. To make the most of EARS, it is crucial to gain an in-depth understanding of how 
routes work in EARS.

Routes connect data source to data destinations, for example you can connect an SQS queue to a Kafka topic. 
Events travelling through such a route can be transformed or filtered according to the configuration details of 
the route.
 
A route consists of a linear sequence of a receiver plugin, followed by zero, one or more filter plugins (the filter chain) 
and a sender plugin. Note that a filter chain can be empty in which case the route simply consists of a receiver 
directly connected to a sender. In either case a route processes events, one at a time in sequence.

The receiver plugin of a route receives events from an event source. The receiver plugin encapsulates all protocol 
specific implementation and configuration details. EARS currently offers receiver plugins for Kafka, Kinesis, SQS, Redis
Pub/Sub and HTTP. The receiver plugin passes each event it receives from its event source to the filter chain. 
Each filter plugin in the filter chain is configured to perform a specific task. EARS offers a small library of highly 
generic and configurable filters to perform various types of payload transformations and filtering out events that do 
not meet predefined criteria such as presence or absence of a certain key value pair etc. After
traveling through the filter chain, events finally arrive at the sender plugin which will then deliver the events to
a data destination. Again EARS offers a set of configurable sender plugins for Kafka, Kinesis, SQS, Redis
Pub/Sub and HTTP.

Note: If needed, you may consider implementing your own sender, receiver or filter plugin in go. To do that you need
to implement the following interfaces. You can also use existing plugin implementations as your guide.
