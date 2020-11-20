# Research

Links to various resources that were similar that helped inform the
project design.


## Code Organization
* https://github.com/golang-standards/project-layout
* https://christine.website/blog/within-go-repo-layout-2020-09-07


## Naming
* https://dave.cheney.net/2019/01/08/avoid-package-names-like-base-util-or-common
* https://github.com/golang/go/wiki/CodeReviewComments#package-names
* https://blog.golang.org/package-names 

## Interfaces
* https://github.com/Evertras/go-interface-examples/tree/master/local-interfaces


## Lifecycle
* https://blog.labix.org/2011/10/09/death-of-goroutines-under-control
* https://dave.cheney.net/2016/12/22/never-start-a-goroutine-without-knowing-how-it-will-stop
* https://github.com/oklog/run
* Uber FX
  * https://godoc.org/go.uber.org/fx
  * https://blog.huyage.dev/posts/simple-dependency-injection-in-go/
  * https://github.com/yagehu/sample-fx-app/blob/master/internal/loggerfx/module.go
  * https://github.com/preslavmihaylov/fxappexample



## Plugin Framework
* [Go Plugins](https://golang.org/pkg/plugin/)
* [BFE Load Balancer](https://github.com/bfenetworks/bfe)
* [OBS Plugins](https://obsproject.com/docs/plugins.html)
* [Markdown "extensions"](https://github.com/yuin/goldmark#custom-parser-and-renderer)


## Data Pipelines
* [Go Blog - Pipelines](https://blog.golang.org/pipelines)
* [Pipe](https://github.com/pipelined/pipe)
* [BFE Load Balancer](https://github.com/bfenetworks/bfe)
* [Go-pipeline](https://whiskybadger.io/post/introducing-go-pipeline/)
* [Multistep](https://github.com/mitchellh/multistep)

## Concurrency
* [Error handling](https://bionic.fullstory.com/why-you-should-be-using-errgroup-withcontext-in-golang-server-handlers/)
* [Sync Pools](https://www.akshaydeo.com/blog/2017/12/23/How-did-I-improve-latency-by-700-percent-using-syncPool/)


## Rate Limiting
* https://github.com/uber-go/ratelimit


## Filters
* [Gojq](https://github.com/itchyny/gojq )


## Publish / Subscriber
* [Watermill.io](https://watermill.io/docs/pub-sub/)
* https://github.com/lileio/pubsub
* [Machinery](https://github.com/RichardKnop/machinery)

## OpenAPI
* https://github.com/deepmap/oapi-codegen


## Testing
* [Log verification](https://github.com/sendgrid/filegetter/blob/master/getter/getter_test.go)
* SQL
  * [In memory SQL](https://github.com/liquidata-inc/go-mysql-server/)
  * https://github.com/proullon/ramsql
  * https://github.com/DATA-DOG/go-sqlmock