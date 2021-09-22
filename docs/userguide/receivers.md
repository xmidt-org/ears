# Receiver Plugin Reference

Receiver plugins receive events from event sources. An EARS route is protocol agnostic, but a
receiver plugin understands the specifics of its event source protocol and encapsulates the 
necessary implementation details. After receiving an event from the event source the receiver
plugin delivers copies of the event to all interested routes. For each event copy, the receiver plugin
waits until the event gets either acknowledged (ack()) or not acknowledged (nack()) by the route or until 
the event times out (typically after 5 seconds). It is then up to the receiver plugin to decide
whether the event has been dealt with or whether the event should be retried. In this sense it is
the receiver plugin's responsibility to orchestrate and oversee any event delivery policies. 

Each receiver plugin instance is determined by its ID, its type, its optional name and its
protocol specific configuration such as topic name or broker endpoint.

```
{
  "receiver": {
    "plugin": "{receiverType}",
    "name": "{optionalReceiverName}",
    "config": {
      "param1": "value1",
      "param2": "value2",
      "...": "...",
    }
  }
}
```

When determining whether two plugin configurations are identical for the purpose of stream sharing or to
find out whether a plugin instance needs updating, EARS calculates a hash over all these parameters and
compares that hash against other hashes.

```
pluginHash = md5(pluginType, pluginName, pluginConfigurationJson)
```

By choosing unique plugin names it is possible to force EARS to use two distinct receiver plugin instances
for two routes with otherwise identical receiver plugin configurations.

## Available Receiver Plugins

* kafka
* kinesis
* sqs
* redis
* http
* debug

### Kafka Receiver Plugin

Example Configuration:

```
{
  "receiver": {
    "plugin": "kafka",
    "name": "myKafkaReceiver",
    "config": {
      "groupId": "myGroup"
    }
  }
}
```

Parameters:

```
type ReceiverConfig struct {
	Brokers             string `json:"brokers,omitempty"`
	Topic               string `json:"topic,omitempty"`
	GroupId             string `json:"groupId,omitempty"`
	Username            string `json:"username,omitempty"` // yaml
	Password            string `json:"password,omitempty"`
	CACert              string `json:"caCert,omitempty"`
	AccessCert          string `json:"accessCert,omitempty"`
	AccessKey           string `json:"accessKey,omitempty"`
	Version             string `json:"version,omitempty"`
	CommitInterval      *int   `json:"commitInterval,omitempty"`
	ChannelBufferSize   *int   `json:"channelBufferSize,omitempty"`
	ConsumeByPartitions bool   `json:"consumeByPartitions,omitempty"`
	TLSEnable           bool   `json:"tlsEnable,omitempty"`
}
```

Default Values:

```
{
	brokers:           "localhost:9092",
	topic:             "quickstart-events",
	groupId:           "",
	username:          "",
	password:          "",
	caCert:            "",
	accessCert:        "",
	accessKey:         "",
	version:           "",
	commitInterval:    1,
	channelBufferSize: 0
}
```

### Kinesis Receiver Plugin

Example Configuration:

```
{
  "receiver": {
    "plugin": "kinesis",
    "name": "myKinesisReceiver",
    "config": {
      "streamName": "ears-demo"
    }
  }
}
```

Parameters:

```
type ReceiverConfig struct {
	StreamName         string `json:"streamName,omitempty"`
	ReceiverPoolSize   *int   `json:"receiverPoolSize,omitempty"`
	AcknowledgeTimeout *int   `json:"acknowledgeTimeout,omitempty"`
	ShardIteratorType  string `json:"shardIteratorType,omitempty"`
}
```

Default Values:

```
{
	streamName:         "",
	receiverPoolSize:   1,
	acknowledgeTimeout: 5,
	shardIteratorType:  "LATEST",
}
```

### SQS Receiver Plugin

Example Configuration:

```
{
  "receiver": {
    "plugin": "sqs",
    "name": "mySqsReceiver",
    "config": {
      "queueUrl": "https://sqs.us-west-2.amazonaws.com/{awsAccountId}/ears-demo",
      "receiverPoolSize": 1,
      "visibilityTimeout": 5
    }
  }
}
```

Parameters:

```
type ReceiverConfig struct {
	QueueUrl            string `json:"queueUrl,omitempty"`
	MaxNumberOfMessages *int   `json:"maxNumberOfMessages,omitempty"`
	VisibilityTimeout   *int   `json:"visibilityTimeout,omitempty"`
	WaitTimeSeconds     *int   `json:"waitTimeSeconds,omitempty"`
	AcknowledgeTimeout  *int   `json:"acknowledgeTimeout,omitempty"`
	NumRetries          *int   `json:"numRetries,omitempty"`
	ReceiverQueueDepth  *int   `json:"receiverQueueDepth,omitempty"`
	ReceiverPoolSize    *int   `json:"receiverPoolSize,omitempty"`
	NeverDelete         *bool  `json:"neverDelete,omitempty"`
}
```

Default Values:

```
{
	queueUrl:            "",
	maxNumberOfMessages: 10,
	visibilityTimeout:   10,
	waitTimeSeconds:     10,
	acknowledgeTimeout:  5,
	numRetries:          0,
	receiverQueueDepth:  100,
	receiverPoolSize:    1,
	neverDelete:         false
}
```

_neverDelete_, when set to true will prevent the SQS receiver from ever deleting messages from the queue. This setting
can be useful when testing with events consumed from a production queue to avoid the risk of message loss.

### Redis Receiver Plugin

Example Configuration:

```
{
  "receiver": {
    "plugin": "redis",
    "name": "myRedisReceiver",
    "config": {
      "channel" : "ears_demo"
    }
  }
}
```

Parameters:

```
type ReceiverConfig struct {
	Endpoint string `json:"endpoint,omitempty"`
	Channel  string `json:"channel,omitempty"`
}
```

Default Values:

```
{
	endpoint: "localhost:6379",
	channel:  "ears",
}
```

### Http Receiver Plugin

Example Configuration:

```
{
    "path" : "mywebhook,
    "method" : POST",
    "port" : 8080
}

```

Parameters:

```
type ReceiverConfig struct {
	Path   string `json:"path"`
	Method string `json:"method"`
	Port   *int   `json:"port"`
}
```

Default Values:

```
{
}
```

### Debug Receiver Plugin

Use this receiver plugin as a data source for debugging purposes. The debug receiver plugin can produce an arbitrary 
number of hardcoded JSON events defined directly in the receiver configuration.

Example Configuration:

```
```

Parameters:

```
type ReceiverConfig struct {
	IntervalMs *int        `json:"intervalMs,omitempty"`
	Rounds     *int        `json:"rounds,omitempty"` // (-1) signifies infinite routes
	Payload    interface{} `json:"payload,omitempty"`
	MaxHistory *int        `json:"maxHistory,omitempty"`
}
```

Default Values:

```
{
	intervalMs: 100,
	rounds:     4,
	payload:    "debug message",
	maxHistory: 100,
}```

_rounds_ defines how many events should be sent.

_intervalMs_ defines how many milliseconds to pause bewteen any two events.

_payload_ is the hardcoded test event payload.

_maxHistory_ defines the number of past events to keep. Begins replacing the oldest event when the buffer is full.




