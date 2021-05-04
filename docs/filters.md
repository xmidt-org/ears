# Filters

## Design Principles

* Syntax should be consistent and concise
* Functionality should be maximally generic and flexible
* Choose default config values wisely so they may often be omitted

## Assumptions And Conventions

* Payloads are either byte arrays or strings. If a payload is a string it will be parsed as JSON into an `interface{}`. Thus the payload will be one of `map[string]interface{}`, `[]interface{}`, `string`, `bool` or `float`. 
* Some filter configs require to specify a particular subset of a deeply nested map. We use dot-delimited path syntax for this purpose, for example `.foo.bar.baz` to reach the value `hello` of the payload `{ "foo" : { "bar" : { "baz" : "hello" }}}`. The root path is defined as `.` or blank string.
* Some filters take data from one sub section of a payload and move them to another sub section of the same payload. By convention we use the configs `FromPath`and `ToPath` for this purpose. If `ToPath` is omitted, then `FromPath` will be used as `ToPath` as well. 
* Events carry two distinct pieces of data objects: payload and metadata and in the future possibly more. Metadata can be used to store temporary data that is produced by one filter and may be consumed by another filter or sender downstream. We use the path prefix `payload.` and `metadata.` to indicate if a filter should operate on metadata or payload. By default we assume a path is used for payloads so the path `.foo` is equivalent to `payload.foo`. 

## Standard Library Of Filters

* match
* decode
* encode
* transform
* hash
* js
* split
* log
* unwrap
* ttl
* trace
* dedup

## Match

### Description

Match an event against an event pattern, a regular expression or both. If a match is found
either let the event pass or drop it. The supported modes are `pattern`, `regex` and `patternregex`.

### Example

Allow events that contain a key `TopicArn` with a string value ending in `_ACCOUNT` or
`_ACCOUNTPRODUCT`.

### Filter Config

```
{
  "plugin": "match",
  "config": {
    "mode": "allow",
    "matcher": "patternregex",
    "pattern": {
      "body" : {
        "TopicArn": ".*_ACCOUNT$|.*_ACCOUNTPRODUCT$"
      }
    }
  }
}
```

## Decode

### Description

Decode entire event payload or metadata, or part of event payload or metadata. Currently only base64 encoding 
is supported.

### Example

Replace the content of the message field with its base64 decoded value.  

### Filter Config

```
{
  "plugin": "decode",
  "name": "useCaseOneRouteDecoder",
  "config": {
    "fromPath": "payload.body.message",
    "toPath" : "payload.body.message",
    "encoding" : "base64" 
  }
}
```

Omitting configs for default behavior we can reduce this to:

```
{
  "plugin": "decode",
  "name": "useCaseOneRouteDecoder",
  "config": {
    "fromPath": ".body.message"
  }
}
```

## Encode

### Description

Encode entire event payload or metadata, or part of event payload or metadata. Currently only base64 encoding
is supported.

### Example

Replace the content of the message field with its base64 encoded value.

### Filter Config

```
{
  "plugin": "encode",
  "config": {
    "fromPath": "payload..body.message",
    "toPath" : "payload.body.message",
    "encoding: " "base64"
  }
}
```

Omitting configs for default behavior we can reduce this to:

```
{
  "plugin": "decode",
  "config": {
    "fromPath": ".body.message"
  }
}
```

## transform

### Description

Basic structural transformations of event payloads and or metadata. Existing fields can be moved around
or filtered. We are using transformation by example syntax for the transformed event and dot-delimited path
syntax to choose from the original event. The filter supports the `ToPath` config which you can use 
to move arbitrary pieces of data between the payload and the metadata section of the event.

### Example

A typical structural transformation of an event payload. The configuration only uses the transformation config,
all other configs use default values.

### Filter Config

```
{
  "plugin": "transform",
  "config": {
    "transformation": {
      "message": {
        "op": "process",
        "payload": {
          "metadata": {
            "type": "{.Message.header.type}",
            "event": "{.Message.header.event}",
            "timestamp": "{.Message.header.timestampMs}",
            "id": "{.Message.body.account.id}"
          },
          "body": {
            "addedAccountProducts": "{.Message.body.addedAccountProducts}",
            "removedAccountProducts": "{.Message.body.removedAccountProducts}"
          }
        }
      },
      "to": {
        "location": "{.Message.body.account.id}",
        "app": "myapp"
      }
    }
  }
},
```

## js

### Description

Arbitrary filtering or transformation operation given in JavaScript source code. Script must return
either `null` (if the event is to be filtered), a single event (original event or modified event), or an
array of events if the event is to be split. The event object must contain a `payload` section and optionally
a `metadata` section. The original event is reachable using the context variable `_.event`. 
Multiline scripts usually look better in YAML representation.

### Example

Filter the event if two array values in the payload are both empty ("nothing to do").

### Filter Config

```
- plugin: js
  config:
    source: |-
      if (_.event.payload.message.payload.body.addedAccountProducts.length == 0 && _.event.payload.message.payload.body.removedAccountProducts.length == 0) {
        return null;
      } else {
        return _.event;
      }
```

## hash

### Description

Calculate a hash over a (subset of) the payload or metadata of an event.
The filter supports the standard configs `FromPath` and `ToPath`.
Supported hash algorithms are fnv, md5, sha1 and sha-256.

### Example

Apply Fowler–Noll–Vo hash modulo 100 to a subsection of the event payload and store the
result in the metadata section for later downstream use in the route. 

### Filter Config

```
{
  "plugin": "hash",
  "config": {
    "fromPath": "payload.to.location",
    "toPath": "metadata.kafka.partition",
    "hashAlgorithm": "fnv"
  }
}
```

or equivalent

```
{
  "plugin": "hash",
  "config": {
    "fromPath": ".to.location",
    "toPath": "metadata.kafka.partition",
    "hashAlgorithm": "fnv"
  }
}
```

## split

### Description

Split an event containing an array in its payload into multiple events. 

### Example

### Filter Config

```
{
  "plugin" : "split",
  "config" : {
    "splitPath" : "."
  }
}
```

Relying on default config values this reduces to:

```
{
  "plugin" : "split"
}
```

## log 

### Description

Logs the current payload and metadata values of the event. For debug purposes only.

### Filter Config

```
{
  "plugin" : "log"
}
```

## unwrap

### Description

Take a subsection of the event and make the root of the event payload, throwing everything else away.
This is a typical requirement when removing an event payload from an envelope. 

### Filter Config

```
{
  "plugin" : "unwrap",
  "config" : {
    "unwrapPath" : ".content"
  }
}

```

## ttl

### Description

Filter an event if it is too old. Requires a valid timestamp to be present in the event payload.
`nanoFactor` is the factor that must be applied to convert the event timestamp into nanoseconds.
The actual `ttl` is given in milliseconds.

### Example

Expire events that arte older than five minutes.

### Filter Config

```
{
  "plugin" : "ttl",
  "config" : {
    "path" : ".content.timestamp",
    "nanoFactor" : 1,
    "ttl" : 300000
  }
}
```

## trace

### Description

Add standard trace information to event payload.

### Filter Config

```
{
  "plugin": "trace",
  "config": {
    "tracePath" : ".tx.traceId"
  }
}
```

## dedup

### Description

Ensure same payload is not sent twice among the most recent N events. Filter uses md5 hashes to store
event signatures. Deduping may be applied to only a subsection of the payload, by default the entire
payload will be considered. Pay extra attention when naming an instance of this filter and consider 
side effects when sharing an instance of this filter type among several routes!

### Filter Config

```
{
  "plugin" : "dedup",
  "config" : {
    "cacheSize" : 1000,
    "dedupPath: : "."
  }
}
```

Or with default values:

```
{
  "plugin" : "dedup"
}
```

## Open Issues

* Keys containing dot `.`
* How to access array elements

