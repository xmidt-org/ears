# Filter Plugins

Filter plugins receive a single event from either a receiver plugin or another filter plugin.
After applying arbitrary filtering logic and / or transformations a filter plugin then forwards
zero, one or more events to a sender plugin or another filter plugin. 

The Filter interface:

```
Filter(evt event.Event) []event.Event
```

Each filter plugin instance is determined by its ID, its type, its optional name and its
filter specific configuration. Often multiple filters are combined to a filter chain.

```
{
  "filterChain": [
    {
      "plugin": "{filterType}",
      "name": "{optionalFilterName}",
      "config": {
        "param1": "value1",
        "param2": "value2",
        "...": "...",
      }
    }
  ]
}
```

EARS comes with a small library of highly generic and configurable filter plugins. If needed you can develop 
your own filter plugin, but you should consider first if your use case can be realized using one or more 
filters form the standard library. There is also a Javascript filter that can be used to implement arbitrary
filter or transformation logic expressed as JavaScript.

## EARS Event Payload and Metadata

Filters operate on events. Therefore, it is important to understand the EARS event model. Each event comes with a 
payload. Payloads are either byte arrays or strings. If a payload is a string it will be parsed as JSON into an `interface{}`.
Thus, the payload will be one of `map[string]interface{}`, `[]interface{}`, `string`, `bool` or `float`. 

Some filter configurations require to specify a particular subset of a deeply nested map. We use dot-delimited path 
syntax for this purpose, for example `.foo.bar.baz` to reach the value `hello` of the payload 
`{ "foo" : { "bar" : { "baz" : "hello" }}}`. The root path is defined as `.` or blank string.

Some filters take data from one sub section of a payload and move it to another sub section of the same payload. 
By convention, we use the configurations `fromPath` and `toPath` for this purpose. If `toPath` is omitted, 
then `fromPath` will also be used as `toPath` as well.


In addition to the payload, an event carries an extra data object called metadata. Metadata can be used to store 
temporary data that is produced by one filter and may be consumed by another filter or sender downstream. 
An example when this can be useful is when you use a _hash_ filter to create a hash that you then later use in 
a _kafka_ sender to select the correct topic partition. In this case the pre-calculated hash needs to be stored
somewhere in the event without polluting its payload. We use the path prefix `payload.` and `metadata.` to indicate 
if a filter should operate on metadata or payload. By default, we assume a path is used for payloads so the path 
`.foo` is equivalent to `payload.foo`. But if you want to access the foo key in the metadata object you have to use
`metadata.foo` as path configuration.

Current limitations: Value keys containing the dot character are currently not supported. Also, an index selector for 
array elements is currently not supported.

## Standard Library Of Filter Plugins

* match
* decode
* encode
* transform
* hash
* regex  
* js
* split
* log
* unwrap
* ttl
* trace

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

### Parameters

```
type Config struct {
	Mode            ModeType    `json:"mode,omitempty"`
	Matcher         MatcherType `json:"matcher,omitempty"`
	Pattern         interface{} `json:"pattern,omitempty"`
	ExactArrayMatch *bool       `json:"exactArrayMatch,omitempty"`
}
```

### Default Values

```
{
	Mode:            "allow",
	Matcher:         "regex",
	Pattern:         "^.*$",
	ExactArrayMatch: true
}
```

Valid values for _mode_  are _allow_ and _deny_.

Valid values for _matcher_ are _pattern_, _regex_ and _patternregex_. The _pattern_ matcher allows you to specify 
a JSON fragment that must be part of the payload for the event to match. The _pattern_ matcher allows * as wildcard 
character for values. The _regex_ matcher allows to specify a single regular expression for the entire payload.
The _patternregex_ matcher combines the two: You specify a JSON fragment and one or more of its values may be 
specified as regular expressions.

_ExactArrayMatch_, if set to true, requires arrays in the pattern to match exactly, meaning same number elements 
and same values. If set to false the payload may contain additional elements in the array not present in the 
pattern. Default is true.

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
            "timestamp": "{.Message.header.timestamp}",
            "id": "{.Message.body.account.id}"
          },
          "body": {
            "addedAccountProducts": "{.Message.body.added}",
            "removedAccountProducts": "{.Message.body.removed}"
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
Supported hash algorithms are _fnv_, _md5_, _sha1_ and _sha-256_.

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

## regex

### Description

Perform arbitrary regular expression transformations.

### Example

Select all digits from _content_ (leftmost match only) and place them into 
_regexedContent_.

### Filter Config

```
{
  "plugin": "regex",
  "config": {
    "fromPath": ".conent",
    "toPath": ".regexedContent",
    "regex": "[0-9]+"
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
    "path" : ".content"
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
    "path" : ".tx.traceId"
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
    "path: : "."
  }
}
```

Or with default values:

```
{
  "plugin" : "dedup"
}
```

