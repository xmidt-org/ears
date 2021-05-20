# Metrics

Metrics are pretty great. This is an explanation of how the XMiDT team uses
[prometheus
metrics](https://pkg.go.dev/github.com/prometheus/client_golang/prometheus). We
use our [touchstone](https://pkg.go.dev/github.com/xmidt-org/touchstone) module
for metrics with uber fx.

## Table of Contents

1. [Definitions](#Definitions)
1. [Summary](#Summary)
1. [Metric Code](#Metric-Code)
1. [Exporting Metrics](#Exporting-Metrics)
1. [Prometheus Deployment](#Prometheus-Deployment)
    1. [Dev](#Dev)
    1. [Prod](#Prod)
1. [Where to Add Metrics](#Where-to-Add-Metrics)
    1. [A Few Tips](#A-Few-Tips)

## Definitions

*Metric*: A metric is a way of measuring a system, feature, or task. All
prometheus metrics are stored as time series data. Different types of metrics
and their explanations can be found
[here](https://prometheus.io/docs/concepts/metric_types/). The metrics in the
prometheus package are all
[Collectors](https://pkg.go.dev/github.com/prometheus/client_golang/prometheus#Collector).

*Registry*: A registry has all metrics registered to it on service startup,
gathering data from them while the service runs.  It is both a
[Registerer](https://pkg.go.dev/github.com/prometheus/client_golang/prometheus#Registerer)
and a
[Gatherer](https://pkg.go.dev/github.com/prometheus/client_golang/prometheus#Gatherer).
The Gatherer is what provides metric data [for http
calls](https://pkg.go.dev/github.com/prometheus/client_golang@v1.10.0/prometheus/promhttp#HandlerFor).

## Summary

Metrics work happens on service startup and throughout the time the service is
running.  Startup work involves creating a registry and registering all metrics
to it.  As the service runs, metrics are updated as things happen. During this
time, the registry gathers data from all known metrics and provides them in a
consumable way.

## Metric Code

Metrics are usually part of objects already doing other business logic. A struct
that updates a set of metrics is not responsible for creating the metrics.
Usually, they are created then provided to the
[struct](https://github.com/xmidt-org/glaukos/blob/16c6eba0b81ec0e1d870bcfa771a10cccd6fc9de/events/codexClient.go#L46).
With uber fx, each measure is created independently with a
[`ProvideMetrics()`](https://github.com/xmidt-org/glaukos/blob/16c6eba0b81ec0e1d870bcfa771a10cccd6fc9de/events/metrics.go#L41)
function. The
[`Measures`](https://github.com/xmidt-org/glaukos/blob/16c6eba0b81ec0e1d870bcfa771a10cccd6fc9de/events/metrics.go#L32-L38)
struct embeds `fx.In`, with each metric tagged.  Then, [`ProvideMetrics()` is
called](https://github.com/xmidt-org/glaukos/blob/16c6eba0b81ec0e1d870bcfa771a10cccd6fc9de/events/provide.go#L67)
as an `fx.Option` somewhere in the application.

Below is some example code explaining the various pieces explained above:
```go
// TODO: add this.
```

## Exporting Metrics

We use the `touchstone` and `touchhttp` packages to handle the wiring and common
code around enabling metrics in our services.  For uber fx applications, calling
[`touchstone.Provide()`](https://pkg.go.dev/github.com/xmidt-org/touchstone@v0.0.3#Provide)
and
[`touchhttp.Provide()`](https://pkg.go.dev/github.com/xmidt-org/touchstone@v0.0.3/touchhttp#Provide)
provides everything you need for setting up a registry and handling a metrics
endpoint.  The documentation explain what each function is providing.

These are the general steps for wiring and exporting metrics:

1. Set up a registry. `touchstone.Provide()` provides this.
1. Register all desired metrics.  This is done by calling various metrics'
   `Provide()` functions as created in your code (see [Metric
   Code](#Metric-Code) for more information).
1. Set up a
   [Handler](https://pkg.go.dev/github.com/xmidt-org/touchstone@v0.0.3/touchhttp#Handler)
   for a metrics endpoint. `touchhttp.Provide()` provides this.
1. [Attach the
   handler](https://github.com/xmidt-org/glaukos/blob/16c6eba0b81ec0e1d870bcfa771a10cccd6fc9de/routes.go#L36)
   to a router or otherwise invoke it to handle your metrics endpoint.  We use
   [`arrange-http`](https://pkg.go.dev/github.com/xmidt-org/arrange@v0.3.0/arrangehttp)
   to [build our
   routers](https://github.com/xmidt-org/glaukos/blob/16c6eba0b81ec0e1d870bcfa771a10cccd6fc9de/main.go#L97)
   from configuration.

## Prometheus Deployment

### Dev

Any manual testing we do while developing is done with a local docker setup.  A few examples of our various docker setups:

- [Glaukos](https://github.com/xmidt-org/glaukos/blob/main/deploy/docker-compose/docker-compose.yml)
- [Svalinn](https://github.com/xmidt-org/codex-deploy/blob/main/deploy/docker-compose/docker-compose.yml)
- [XMiDT](https://github.com/xmidt-org/xmidt/blob/master/deploy/docker-compose/docker-compose.yml)

A prometheus instance is included in the docker setup and can be reached at
`http://localhost:9090/`. All services involved have exposed their metric port,
and Prometheus is
[configured](https://github.com/xmidt-org/glaukos/blob/main/deploy/docker-compose/docFiles/prometheus.yml)
to scrape them.  Then, while testing, the developer has visibility into metric
data in addition to output.

### Prod

For production, prometheus is deployed and configured to find services to scrape using consul.

## Where to Add Metrics

Generally, we add metrics to track these types of things:
- **HTTP requests**: both [incoming
  (server)](https://pkg.go.dev/github.com/xmidt-org/touchstone@v0.0.3/touchhttp#ServerBundle)
  and [outgoing
  (client)](https://pkg.go.dev/github.com/xmidt-org/touchstone@v0.0.3/touchhttp#ClientBundle);
  tracking the size and duration of a request as well as the response code.  Our
  services share some general metrics but often implement their own based on
  what is being done - a common one is counting retries for client requests.
- **Auth**: tracking the
  [result](https://github.com/xmidt-org/bascule/blob/c011b128d6b95fa8358228535c63d1945347adaa/basculehttp/metrics.go#L49)
  of auth middleware on http requests, including
  [why](https://github.com/xmidt-org/bascule/blob/c011b128d6b95fa8358228535c63d1945347adaa/basculehttp/errorResponseReason.go#L24-L34)
  a request was considered unauthorized.  With capability checking, we have even
  [more
  data](https://github.com/xmidt-org/bascule/blob/c011b128d6b95fa8358228535c63d1945347adaa/basculechecks/metrics.go#L73-L82),
  showing the client id and partner in addition to the result.
- **Asynchronous work**: 
  - if the response for an http request has been sent but work is still being
    done, the result of that work should be captured in a metric. Often, we have
    failure cases post-response that result in dropping the object we are
    working with, so we [capture the
    count](https://github.com/xmidt-org/glaukos/blob/16c6eba0b81ec0e1d870bcfa771a10cccd6fc9de/eventmetrics/queue/metrics.go#L65-L71)
    and the reason for failure.
  - often, asynchronous work involves some sort of queue/channel, so we track
    the [queue
    depth](https://github.com/xmidt-org/glaukos/blob/16c6eba0b81ec0e1d870bcfa771a10cccd6fc9de/eventmetrics/queue/metrics.go#L51-L56).
    There may also be a limited number of workers for the queue, in which case
    we track the [current number of
    workers](https://github.com/xmidt-org/caduceus/blob/55bce1ccebe0312e1477ddb8a5c58203211b3214/metrics.go#L129-L132).
- **Complex logic**: this varies by what is trying to be accomplished.  Some common motivations for adding metrics:
  - if an object being used has a type or enum, there may be a counter with the
    type label, showing the breakdown of what types are being worked with.
  - if there is something long lived (ie connections), a counter of that and any
    easily-grouped metadata is helpful.

### A Few Tips
  - Generally, greater insight into failures is common.  Successes are great and
    are counted, but usually we want the most data into things that go wrong.
  - [Labels](https://prometheus.io/docs/practices/naming/#labels) add
    cardinality to metrics.  The cardinality of a metric is the cardinality
    (number of possible values) for each label of that metric multiplied
    together.  The reason or "why" of a failure in the context of metrics is
    more like a category of failure rather than something as detailed as an
    error message. Enums don't have to be the only label values, but anything
    that can grow without bounds should be considered carefully before being
    added as a metric label value.
