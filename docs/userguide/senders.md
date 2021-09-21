# Sender Plugin Referecne

Each sender plugin instance is determined by its ID, its type, its optional name and its 
protocol specific configuration such as topic name or broker endpoint.

```
{
  "sender": {
    "plugin": "{sednerType}",
    "name": "{optionalSenderName}",
    "config": {
      "param1": "value1",
      "param2": "value2",
      "...": "...",
    }
  }
}
```

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
	Brokers:             "localhost:9092",
	Topic:               "quickstart-events",
	Partition:           -1,
	ChannelBufferSize:   0,
	Username:            "",
	Password:            "",
	CACert:              "",
	AccessCert:          "",
	AccessKey:           "",
	Version:             "",
	SenderPoolSize:      1,
	PartitionPath:       "",
	DynamicMetricLabels: []
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
    StreamName:          "",
    MaxNumberOfMessages: 1
    SendTimeout:         1
}
```

### SQS Sender Plugin

Example Configuration:

Parameters:

Default Values:

### Kafka Sender Plugin

Example Configuration:

Parameters:

Default Values:

### Kafka Sender Plugin

Example Configuration:

Parameters:

Default Values:

### Kafka Sender Plugin

Example Configuration:

Parameters:

Default Values:

### Kafka Sender Plugin

Example Configuration:

Parameters:

Default Values:

