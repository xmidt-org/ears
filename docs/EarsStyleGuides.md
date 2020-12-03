# (G)EARS Style Guides
A programming style guide for (but not limit to) the EARS project.

## Golang Style Guide
Coding standard for Go programming language

### 1.0 Project layout
For the project layout, we are following [this layout guide](https://github.com/golang-standards/project-layout).

The basic ones are:
* `/cmd`: main applications for the project
* `/internal`: private code that you don't want others to use
* `/pkg`: library code that external applications can use

### 2.0 Source file

#### 2.1 File name
Should we camelCase filenames?

#### 2.2 Open source
Open source file should contain licence information on the top of the file before package and import

#### 2.3 Formatting
All source file must be formatted using `go fmt`

##### Long line
Do we want to impose limits on the number of characters per line?

### 3.0 Naming
Golang's convention is [MixedCaps](https://golang.org/doc/effective_go.html#mixed-caps), which is essentially either PascalCase or camelCase.

What about abbreviations in camelCase?
Google's [Java camelCase style](https://google.github.io/styleguide/javaguide.html#s5.3-camel-case)

#### 3.1 Package name
According to [golang](https://blog.golang.org/package-names): Good package names are short and clear. They are lower case, with no under_scores or mixedCaps. They are often simple nouns.

#### 3.2 Struct, interface and constants
Struct, interface, and constants should be in PascalCase.

#### 3.3 Variables and functions
Variables and functions should follow the golang conventions where public ones are in PascalCase, and private/local ones are in camelCase.

### 4.0 Programming practices

#### 4.1 Context
golang context is normally associated with a transaction such as processing an HTTP request or an incoming event. In these use-cases, there is a context per transaction.

* A context should only carry readonly information for logging, tracing, metrics, and/or API receipt. 
* Application logic should <strong>NOT</strong> depend on values in context.
* Context enables deadline and cancellation. Any slow processing function and libraries should listen to the [context Done channel](https://www.sohamkamani.com/golang/2018-06-17-golang-using-context-cancellation/) for any request/event cancellation.

#### 4.2 Struct instantiation
Simple struct without default values can be instantiated directly. A good example is error struct.

A struct with private fields should be instantiated with a new method such as `NewMyObject(...)`. Some examples are readers/writers, 

Complicated/big struct such as a framework or a manager with a lot of configurations should consider [functional options](https://dave.cheney.net/2014/10/17/functional-options-for-friendly-apis)

#### 4.3 Interfaces
TBD

#### 4.4 Errors
Avoid handling errors by checking the `Error()` string whenever possible. Instead, use `As()` method for checking the error type. The `Error()` method should only be used for informational purpose such as logging or as a response.

Explicitly create an error struct for an error type instead of creating generic errors that can only be differentiated with `Error()` string. 

For example, do this:
```go
//define your error type
type MyError {
    Msg string
}

//instantiate the error
return &MyError{"something is wrong"}

//Checking
var myErr *MyError
if err.As(&myErr)
```

Avoid this:
```go
//generic error type, not useful
return &fmt.Errorf("something is wrong")

//comparing the error string directly. Error prone and hard to maintain
if err.Error() == "something is wrong"
```

What about this?
```go
SomethingIsWrong = errors.New("something is wrong")

return SomethingIsWrong

//checking
if err == SomethingIsWrong
```

##### Error Wrapping
When creating a new error struct, wrap the source error only if the consumer of this error need access to the source error or need to type-check based on the source error. If you decide to wrap the source error, make sure to implement the `Unwrap()` method.

[Further reading on errors](https://blog.golang.org/go1.13-errors)

#### 4.5 Panics
TBD

#### 4.6 Unit tests
TBD

### 5.0 External libraries

#### 5.1 viper and cobra
For ingestions of commandline arguments, configuration files, and/or environmental variables:
* [Viper](https://github.com/spf13/viper)
* [Cobra](https://github.com/spf13/cobra)

#### 5.2 zerolog
Fast and structured logging utility. Compatible with go `context`
* [Zerolog](https://github.com/rs/zerolog)

#### 5.3 uber/fx
Application lifecycle and dependency injection framework
* [Uber/fx](https://github.com/uber-go/fx)

#### 5.4 gorrilla mux
HTTP router/multiplexer for building Rest APIs
* [Gorrilla/mux](https://github.com/gorilla/mux)

## Logging Guide

### 1.0 Reserved Keys
The following data key/values generated automatically. Avoid using these keys to avoid conflicts:

|Key|Description|
|-----|---------|
|app.id| name/ID of the microservice|
|app.imageFullId| |
|app.imageName| container image name|
|app.version| (only in elastic search) |
|appVersion| (only in splunk) |
|cluster.id| |
|cluster.type| dev, qa, sandbox, perf, tps, prod, etc |
|cluster.location| |
|host.id| |
|host.ip| |
|host.name| |
|log.filterChain| |
|log.hostChain| |
|log.id| |
|log.source| stdout, stderr, etc|
|log.timestamp| |

### 2.0 Key/Value Format

Logging keys may contain multiple segments separated by dot (for example, stat.key). Each segment should be in camelCase.

For each distinct logging key, please keep its value type consistent. Do <strong>NOT</strong> mix the value type.

For string value type, it should be in camelCase. The only exception is the `message` field which may contain human readable message sentences.

### 3.0 Common Logging Fields

| Key | Type | Description |
|-----|------|-------------|
|log.level| string | debug, info, warn, error |
|gears.app.id| string | application/tenant ID |
|op| string | operation. For example, createUser API may have `op=createUser` |
|action| string | step in the operation. For example, createUser may have `action=updateDB` |
|error| string | error message, usually from the `Error()` method |
|message| string | human readable message |
|stat.key| string | the name of the stats that we are tracking |
|stat.count| long | the stats count if applicable | 
|stat.time| long | the stats duration in milliseconds if applicable |
|location| string | basic partition unit in gears eco system |

## Configuration Guide
Gears applications leverage [Viper/Cobra](#5.1-viper-and-cobra) for configurations. The application usually takes a configuration file through a commandline flag or an environmental variable.
* `--config` is typically used for giving configuration file at command line
* `GEARS_CONFIG` is typically used for giving configuration file through environmental variable

Configurations in config file can be overwritten with:
* environmental variables???
* command line arguments

### 1.0 Configuration file

#### 1.1 Format
An example of a configuration file
```yaml
ears:
  logLevel: info
  numWokers: 40
  
  api:
    port: 8080
    timeout: 10 
```

* File must be in yaml format
* Keys are in hierarchical orders. 
* The root level key should always be the application name.
* Each level of the key name should be in camelCase
