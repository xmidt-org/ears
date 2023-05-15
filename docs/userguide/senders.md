# Sender Plugin Reference

Sender plugins deliver events to event destinations. An EARS route is protocol agnostic, but a
sender plugin understands the specifics of its event destination protocol and encapsulates the
necessary implementation details.

Each sender plugin instance is determined by its ID, its type, its optional name and its 
protocol specific configuration such as topic name or broker endpoint.

```
{
  "sender": {
    "plugin": "{senderType}",
    "name": "{optionalSenderName}",
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

By choosing unique plugin names it is possible to force EARS to use two distinct sender plugin instances
for two routes with otherwise identical sender plugin configurations.

## Secrets Management

Some senders require authentication information in the form of access tokens etc. There are several reasons
why you would not want such tokens to be part of the route configuration itself: (1) Large access tokens make the
route less readable, (2) every key rotation will require redployment of the route and change the plugin hash
and (3) there are security concerns about secrets then being visible in the response of the GetRoutes() APIs.
Therefore, EARS supports a simple layer of indirection where you can refer to a secret by a short name instead
placing the secret itself directly into the route. For this purpose EARS supports the following syntax for
selected plugin configurations: `secret://kafka.accessCert`. Currently, EARS only supports using the service
configuration file ears.yaml as a data store for secrets. Your ears.yaml must contain the following section:

```
ears:
    secrets:
        myorg:
            myapp:
                kafka:
                    accessCert: |-
                        ...
                    accessKey: |-
                        ...
```

Note that secrets are isolated by org ID and app ID for multi tenancy.

## Available Sender Plugins

* kafka
* kinesis
* sqs
* redis
* http
* debug

### Kafka Sender Plugin

Example Configuration:

```
{
  "sender": {
    "plugin": "kafka",
    "name": "myKafkaSender",
    "config": {
      "brokers": "kafkabroker:16383",
      "topic": "kafkatopic",
      "caCert": "secret://kafka.caCert",
      "accessCert": "secret://kafka.accessCert",
      "accessKey": "secret://kafka.accessKey"
    }
  }
}
```

Parameters:

```
type SenderConfig struct {
	Brokers             string               `json:"brokers,omitempty"`
	Topic               string               `json:"topic,omitempty"`
	Partition           *int                 `json:"partition,omitempty"`
	PartitionPath       string               `json:"partitionPath,omitempty"` 
	Username            string               `json:"username,omitempty"`
	Password            string               `json:"password,omitempty"`
	CACert              string               `json:"caCert,omitempty"`
	AccessCert          string               `json:"accessCert,omitempty"`
	AccessKey           string               `json:"accessKey,omitempty"`
	Version             string               `json:"version,omitempty"`
	ChannelBufferSize   *int                 `json:"channelBufferSize,omitempty"`
	TLSEnable           bool                 `json:"tlsEnable,omitempty"`
	SenderPoolSize      *int                 `json:"senderPoolSize,omitempty"`
	DynamicMetricLabels []DynamicMetricLabel `json:"dynamicMetricLabel,omitempty"`
}
```

Default Values:

```
{
	brokers:             "localhost:9092",
	topic:               "quickstart-events",
	partition:           -1,
	channelBufferSize:   0,
	username:            "",
	password:            "",
	caCert:              "",
	accessCert:          "",
	accessKey:           "",
	version:             "",
	senderPoolSize:      1,
	partitionPath:       "",
	dynamicMetricLabels: []
}
```

If _PartitionPath_ is set, it is used to look up partition information from the event (payload or metadata), rather than using
a hard coded _Partition_. Default value is -1 for random partition.

### Kinesis Sender Plugin

Example Configuration:

```
{
    "sender": {
        "plugin": "kinesis",
        "name": "myKinesisSender",
        "config": {
            "streamName": "ears-demo"
        }
    }
}
```

Parameters:

```
type SenderConfig struct {
    StreamName          string `json:"streamName,omitempty"`
    MaxNumberOfMessages *int   `json:"maxNumberOfMessages,omitempty"`
    SendTimeout         *int   `json:"sendTimeout,omitempty"`
}
```

Default Values:

```
{
    streamName:          "",
    maxNumberOfMessages: 1
    sendTimeout:         1
}
```

### SQS Sender Plugin

Example Configuration:

```
{
  "sender": {
    "plugin": "sqs",
    "name": "mySqsSender",
    "config": {
      "queueUrl": "https://sqs.us-west-2.amazonaws.com/myawsaccount/earsdemo"
    }
  }
}
```

Parameters:

```
type SenderConfig struct {
	QueueUrl            string `json:"queueUrl,omitempty"`
	MaxNumberOfMessages *int   `json:"maxNumberOfMessages,omitempty"`
	SendTimeout         *int   `json:"sendTimeout,omitempty"`
	DelaySeconds        *int   `json:"delaySeconds,omitempty"`
}
```

Default Values:

```
{
	queueUrl:            "",
	maxNumberOfMessages: 10,
	sendTimeout:         1,
	delaySeconds:        0
}
```

### Redis Sender Plugin

Example Configuration:

```
{
  "sender": {
    "plugin": "redis",
    "name": "myRedisSender",
    "config": {
      "channel" : "ears_demo"
    }
  }
}
```

Parameters:

```
type SenderConfig struct {
	Endpoint string `json:"endpoint,omitempty"`
	Channel  string `json:"channel,omitempty"`
}
```

Default Values:

```
{
	endpoint: "localhost:6379",
	channel:  "ears"
}
```

### Http Sender Plugin

Example Configuration:

```
{
    "url" : "http://someendpoint,
    "method" : POST"
}
```

Parameters:

```
type SenderConfig struct {
	Url    string `json:"url"`
	Method string `json:"method"`
}
```

Default Values:

```
{
}
```

### Discord Sender Plugin

Example Configuration:

```
{
  "botToken": "xxxxxxxxxxxxxxxxxxxxxxxxxx",
}
```

Parameters:

```
type SenderConfig struct {
  BotToken        string `json:"botToken"`
}
```

Default Values:

```
{
  "botToken": ""
}
```

### Debug Sender Plugin

Use this sender plugin as a data sink for debugging purposes. The debug sender plugin can print payloads to stdout
or stderr or throw them away.

Example Configuration:

```
{
  "sender": {
    "plugin": "debug",
    "name": "myDebugSender",
    "config": {
      "destination": "stdout",
      "maxHistory": 100
    }
  }
}
```

Parameters:

```
type SenderConfig struct {
	Destination DestinationType `json:"destination,omitempty"`
	MaxHistory *int `json:"maxHistory,omitempty"`
}
```

Default Values:

```
{
	destination: "devnull",
	maxHistory:  0,
}
```

_maxHistory_ defines the number of past events to keep. Begins replacing the oldest event when the buffer is full.

_destination_ should be one of _devnull_, _stdout_, _stderr_



