# Ears Plugins Proposal

![](Ears%20Plugins.png)

Legends:
* Blue: Input Plugins
* Yellow: Output Plugins
* Green: Event Processing Plugins

## Proposal

An event pipeline consists of an input plugin, 0 or more event processing plugins, and an output plugin.

Users configure event pipelines through EARS APIs

Users configure plugins through EARS APIs. The plugin configurations are opaque to core EARS.

Plugins has input and output event type to allow EARS validate if a pipeline can be constructed.

### Plugin Interface

```
type EarsPlugin interface {
    GetInputEventType() string
    GetOutputEventType() string
    SetConfig(config string) error
    ... other stuffs ...
}
```

For MVP, the event types are:
* any: the plugin may accept/emit events in any format
* json: events are in json format
* binary: events are in binary format

When EARS validates a pipeline, the two plugins connecting to each other must have matching event type such that the first plugin's output event type must match the second plugin's input event type. The only exception is `any` which can match any event type.

The event types are baked into the plugin code. They are strictly read-only properties for EARS.

### Pipeline configuration example
The configuration format is for illustration purpose. They are not fully hashed out. Please focus on the idea behind it rather than the actual formatting.

Please note a few things:
* the configuration in each plugin are plugin specific. They can be in any formats. EARS does not dictate their formats.
* Kinesis input plugin has an output event type of `any`
* Json_splitter, json_filter, and json_ts_filter's input and output event types are `json`.
* Gears output plugin has an output event type of `json`

#### History Service Event Pipeline
```
{
  "kinesisIn": {           //key is an abitrary name
     "type": "input",      //this is an input plugin
     "plugin": "kinesis",  //refer to the so file name?
     "configuration": {
        ...
     },
     "outputTo": "eventSplitter"
  },
  "eventSplitter": {
     "type": "process",          //this is a process plugin
     "plugin": "json_splitter",  
     "configuration": {
        "arrayKey": ""
     },
     outputTo": "evenFilter"
  },
  "eventFilter": {
    "type": "process",
    "pluging": "json_filter",
    "configuration": {
      "include": {
        "mediaType": "zone/door"
      },
      ...
    }
  },
  "tsFilter": {
    "type": "process",
    "pluging": "json_ts_filter",
    "configuration": {
      "tsField": "event.ts",
      "threshold": 60
    }
  },
  "gearsOut": {
    "type": "output",   //this is an output plugin
    "plugin": "gears",
    "configuration": {
      "location": "event.location",
      "kafka_brokers": "...",
      ...
    }
  }
}   
```
