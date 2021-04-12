# SQS Receiver And Sender Plugin

## Design Notes

* support batch operations for send, receive, delete
* independent threads for send, receive and delete for reduced throughput volatility
* added metadata field to event struct to allow keeping SQS message metadata attached to event (message id, receipt handle etc.)
* use SQS receive count to determine if maximum number of retries has been reached
* clean shutdown of receiver and sender loop when deleting or updating route
* dead letter queue not supported for now
* need to figure out how to access event from Ack() and Nack()
* for now using plugin-local logger, consider CreateEventContext helper function to start trace in receiver
* measure start time and total event code at plugin level for throughput measurements

## Code Review

What I would like to write...

```
e, err := event.New(ctx, payload, event.WithMetadata(*message), event.WithAck(
func() {
    msg := e.Metadata().(sqs.Message) // get metadata associated with this event
    logger.Info().Str("op", "SQS.Receive").Msg("processed message " + (*msg.MessageId))
    var entry sqs.DeleteMessageBatchRequestEntry
    entry = sqs.DeleteMessageBatchRequestEntry{Id: msg.MessageId, ReceiptHandle: msg.ReceiptHandle}
    entries <- &entry
},
func(err error) {
    msg := e.Metadata().(sqs.Message) // get metadata associated with this event
    logger.Error().Str("op", "SQS.Receive").Msg("failed to process message " + (*msg.MessageId) + ": " + err.Error())
    // a nack below max retries - this is the only case where we do not delete the message yet
}))
```

What I ended up writing instead...

```
e, err := event.New(ctx, payload, event.WithMetadata(*message))
if err != nil {
    logger.Error().Str("op", "SQS.Receive").Msg("cannot create event: " + err.Error())
    return
}
err = e.SetAck(
    func() {
        var entry sqs.DeleteMessageBatchRequestEntry
        msg := e.Metadata().(sqs.Message) // get metadata associated with this event
        logger.Info().Str("op", "SQS.Receive").Msg("processed message " + (*msg.MessageId))
        entry = sqs.DeleteMessageBatchRequestEntry{Id: msg.MessageId, ReceiptHandle: msg.ReceiptHandle}
        entries <- &entry
    },
    func(err error) {
        msg := e.Metadata().(sqs.Message) // get metadata associated with this event
        logger.Error().Str("op", "SQS.Receive").Msg("failed to process message " + (*msg.MessageId) + ": " + err.Error())
        // a nack below max retries - this is the only case where we do not delete the message yet
    })
```

## Receive Config Parameters 

```
"ReceiverConfig": {
    "type": "object",
    "additionalProperties": false,
    "properties": {
        "queueUrl": {
            "type": "string"
        },
        "maxNumberOfMessages": {
            "type": "integer", 
            "minimum": 1,
            "maximum": 10
        },
        "visibilityTimeout": {
            "type": "integer", 
            "minimum": 1
        },
        "waitTimeSeconds": {
            "type": "integer", 
            "minimum": 1
        },
        "acknowledgeTimeout": {
            "type": "integer", 
            "minimum": 1
        },
        "numRetries": {
            "type": "integer", 
            "minimum": 0
        },
        "receiverQueueDepth": {
            "type": "integer", 
            "minimum": 0,
            "maximum": 1000
        }
    },
    "required": [
        "queueUrl"
    ],
    "title": "ReceiverConfig"
}
```

## Send Config Parameters

```
"SenderConfig": {
    "type": "object",
    "additionalProperties": false,
    "properties": {
        "queueUrl": {
            "type": "string"
        },
        "batchSize": {
            "type": "integer", 
            "minimum": 1,
            "maximum": 10
        }
    },
    "required": [
        "queueUrl"
    ],
    "title": "SenderConfig"
}
```

## Tests

### Send, Receive & Delete Three Messages

```
bwolf200@cacsvml-18427:~/go/src/github.com/xmidt-org/ears/pkg/plugins/sqs$ go test -run TestSQSSenderReceiver
{"log.level":"info","op":"SQS.Send","batchSize":3,"sendCount":0,"message":"send message batch"}
{"log.level":"info","op":"SQS.Receive","message":"wait for receive done"}
{"log.level":"info","op":"SQS.Receive","batchSize":1,"message":"processed message 1a4072b8-7a50-4817-9be7-88bbfc7bedaa"}
{"log.level":"info","op":"SQS.Receive","batchSize":1,"message":"processed message d46a99bf-2447-4700-b91e-495ecdcec44d"}
{"log.level":"info","op":"SQS.Receive","batchSize":1,"message":"deleted message 1a4072b8-7a50-4817-9be7-88bbfc7bedaa"}
{"log.level":"info","op":"SQS.Receive","batchSize":1,"message":"processed message 6d18d5f7-59ea-420b-a273-31a06668630b"}
{"log.level":"info","op":"SQS.Receive","batchSize":1,"message":"deleted message d46a99bf-2447-4700-b91e-495ecdcec44d"}
{"log.level":"info","op":"SQS.Receive","batchSize":1,"message":"deleted message 6d18d5f7-59ea-420b-a273-31a06668630b"}
{"log.level":"info","op":"SQS.Receive","elapsedMs":2004,"deleteCount":3,"receiveCount":3,"receiveThroughput":1,"deleteThroughput":1,"message":"receive done"}
PASS
ok  	github.com/xmidt-org/ears/pkg/plugins/sqs	3.415s
```

## Throughput

Using batch operations for receive, delete and send improves throughput
compared to trivial implementation. Independent threads for receive and 
delete further increases throughput. Throughput is balanced.

### Batch Size 10

`curl -X POST http://localhost:3000/ears/v1/routes --data @sqsRouteBatch10.json`

```
{
  "id": "sqs102",
  "orgId": "comcast",
  "appId": "xfi",
  "userId": "boris",
  "name": "sqsRoute",
  "receiver": {
    "plugin": "sqs",
    "name": "mySqsReceiver",
    "config": {
      "queueUrl": "https://sqs.us-west-2.amazonaws.com/447701116110/earsSender"
    }
  },
  "sender": {
    "plugin": "sqs",
    "name": "mySqsSender",
    "config": {
      "queueUrl": "https://sqs.us-west-2.amazonaws.com/447701116110/earsSender"
    }
  },
  "deliveryMode": "whoCares"
}
```

```
Message Size: 10 bytes

Receive Count: 621
Receive Throughput: 25 msg/sec

Delete Count: 610
Delete Throughput: 25 msg/sec

Send Count: 597
Send Throughput: 24 msg/sec
```

### Batch Size 1

`curl -X POST http://localhost:3000/ears/v1/routes --data @sqsRouteBatch1.json`

```
{
  "id": "sqs102",
  "orgId": "comcast",
  "appId": "xfi",
  "userId": "boris",
  "name": "sqsRoute",
  "receiver": {
    "plugin": "sqs",
    "name": "mySqsReceiver",
    "config": {
      "queueUrl": "https://sqs.us-west-2.amazonaws.com/447701116110/earsSender",
      "maxNumberOfMessages": 1
    }
  },
  "sender": {
    "plugin": "sqs",
    "name": "mySqsSender",
    "config": {
      "queueUrl": "https://sqs.us-west-2.amazonaws.com/447701116110/earsSender",
      "batchSize": 1
    }
  },
  "deliveryMode": "whoCares"
}
```

```
Message Size: 10 bytes

Receive Count: 579
Receive Throughput: 19 msg/sec

Delete Count: 576
Delete Throughput: 19 msg/sec

Send Count: 564
Send Throughput: 19 msg/sec
```


