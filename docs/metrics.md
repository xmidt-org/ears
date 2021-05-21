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
gathering data from them while the service runs. It is both a
[Registerer](https://pkg.go.dev/github.com/prometheus/client_golang/prometheus#Registerer)
and a
[Gatherer](https://pkg.go.dev/github.com/prometheus/client_golang/prometheus#Gatherer).
The Gatherer is what provides metric data [for http
calls](https://pkg.go.dev/github.com/prometheus/client_golang@v1.10.0/prometheus/promhttp#HandlerFor).

## Summary

Metrics work happens on service startup and throughout the time the service is
running. Startup work involves creating a registry and registering all metrics
to it. As the service runs, metrics are updated as things happen. During this
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
struct embeds `fx.In`, with each metric tagged. Then, [`ProvideMetrics()` is
called](https://github.com/xmidt-org/glaukos/blob/16c6eba0b81ec0e1d870bcfa771a10cccd6fc9de/events/provide.go#L67)
as an `fx.Option` somewhere in the application.

Below is some example code explaining the various pieces summarized above:
```go
// in metrics.go

const(
  // ResultLabel tells us how our example did. It is helpful as a const, so
  // the same string isn't used in multiple places. It only needs to be
  // exported if it is needed outside of the package.
  ResultLabel = "result"
)

// Result label values
const (
  SuccessResult = "success"
  FailureResult = "failure"
)

// Measures contains the metrics needed for our example struct.
type Measures struct {
  fx.In
  ExampleCount *prometheus.CounterVec `name:"example_count"`
}

// ProvideMetrics builds the example-related metrics and makes them available
// to the container.
func ProvideMetrics() fx.Option {
  return fx.Options(
    // touchstone has already been provided, so using CounterVec() allows us to
    // have touchstone manage registering the metric.
    touchstone.CounterVec(
      prometheus.CounterOpts{
        Name: "example_count", // name should match the tag in Measures.
        Help: "Number of times something happens in our example",
      },
      ResultLabel,
      // any number of labels can be added here, but they all need to be given
      // values when this metric is used.
    ),
  )
}
```

```go
// in example.go

// Example adds to its internal counter, resetting on overflow. This is not
// concurrency safe.
type Example struct {
  count int
  max int
  M Measures
}

// Inc adds to the internal counter in Example. If the count hits the maximum,
// an error is returned and the counter is reset.
func (e *Example) Inc() error {
  if e.count >= e.max {
    e.count = 0
    e.M.ExampleCount.With(prometheus.Labels{ // Labels is just a map.
      ResultLabel: FailureResult,
    }).Add(1.0)
    return errors.New("overflow")
  }
  e.count++
  e.M.ExampleCount.With(prometheus.Labels{ // Labels is just a map.
    ResultLabel: SuccessResult,
  }).Add(1.0)
  return nil
}

// NewExample create an example with the given measures and maximum.
func NewExample(m Measures, max int) (*Example, error) {
  if max < 1 {
    return nil, errors.New("max must be greater than 1")
  }
  return &Example{
    max: max,
    M: m,
  }, nil
}
```

## Exporting Metrics

We use the `touchstone` and `touchhttp` packages to handle the wiring and common
code around enabling metrics in our services. For uber fx applications, calling
[`touchstone.Provide()`](https://pkg.go.dev/github.com/xmidt-org/touchstone@v0.0.3#Provide)
and
[`touchhttp.Provide()`](https://pkg.go.dev/github.com/xmidt-org/touchstone@v0.0.3/touchhttp#Provide)
provides everything you need for setting up a registry and handling a metrics
endpoint. The documentation explains what each function is providing.

These are the general steps for wiring and exporting metrics:

1. Set up a registry. `touchstone.Provide()` provides this.
1. Register all desired metrics. This is done by calling various metrics'
   `Provide()` functions as created in your code (see [Metric
   Code](#Metric-Code) for more information).
1. Set up a
   [Handler](https://pkg.go.dev/github.com/xmidt-org/touchstone@v0.0.3/touchhttp#Handler)
   for a metrics endpoint. `touchhttp.Provide()` provides this.
1. [Attach the
   handler](https://github.com/xmidt-org/glaukos/blob/16c6eba0b81ec0e1d870bcfa771a10cccd6fc9de/routes.go#L36)
   to a router or otherwise invoke it to handle your metrics endpoint. We use
   [`arrange-http`](https://pkg.go.dev/github.com/xmidt-org/arrange@v0.3.0/arrangehttp)
   to [build our
   routers](https://github.com/xmidt-org/glaukos/blob/16c6eba0b81ec0e1d870bcfa771a10cccd6fc9de/main.go#L97)
   from configuration.

An example that expands on the code provided in [Metric Code](#Metric-Code) is shown below:

```go
// in main.go

const (
  applicationName = "example"
)

func main() {
  // setup needed to use viper to get configuration.
  v, err := setupViper()
  if err != nil {
    fmt.Fprintln(os.Stderr, err)
    os.Exit(1)
  }
  app := fx.New(
    arrange.ForViper(v), // metrics related configuration is needed, so viper is too
    fx.Provide(
      // read in prometheus configuration from the config file
      arrange.UnmarshalKey("prometheus", touchstone.Config{}),
    ),
    touchstone.Provide(), // provides the metric registry
    touchhttp.Provide(),  // provides the metric handler and some other http metric middleware
    ProvideMetrics(),     // the metrics for Example
    arrangehttp.Server{
      Name: "server_metrics",  // for annotating our mux router
      Key:  "servers.metrics", // key used in the config file
    }.Provide(), // provides the router for metrics
    fx.Provide(
      NewExample,
      // provide the max number. In a real program, this would be annotated.
      func() int {
        return 10
      },
    ),
    fx.Invoke(
      handleMetricEndpoint,
      // this is fine for an example, but long running goroutines should have
      // an exit strategy.
      func(e *Example) {
        go func() {
          for {
            // errors shouldn't be thrown away either.
            _ = e.Inc()
            time.Sleep(time.Second)
          }
        }()
      },
    ),
  )

  // general app stuff below - not a concern metrics-wise.
  if err := app.Err(); err == nil {
    app.Run()
  } else if errors.Is(err, pflag.ErrHelp) {
    return
  } else {
    fmt.Fprintln(os.Stderr, err)
    os.Exit(2)
  }
}

func setupViper() (*viper.Viper, error) {
  v := viper.New()

  v.SetConfigName(applicationName)
  v.AddConfigPath(fmt.Sprintf("/etc/%s", applicationName))
  v.AddConfigPath(fmt.Sprintf("$HOME/.%s", applicationName))
  v.AddConfigPath(".")
  err := v.ReadInConfig()
  if err != nil {
    return v, fmt.Errorf("failed to read config file: %w", err)
  }

  return v, nil
}
```

```go
// in routes.go

type MetricRouterIn struct {
  fx.In
  Router  *mux.Router `name:"server_metrics"`
  Handler touchhttp.Handler
}

func handleMetricEndpoint(in MetricRouterIn) {
  in.Router.Handle("/metrics", in.Handler).Methods("GET")
}
```

Sending a request to the metrics endpoint would result in a response that
includes our metric:

```
# HELP namespace_subsystem_example_count Number of times something happens in our example
# TYPE namespace_subsystem_example_count counter
namespace_subsystem_example_count{result="failure"} 11
namespace_subsystem_example_count{result="success"} 116
```

There will be additional metrics included in the response that are included as a
part of the golang prometheus client and our metric handler. The metrics are
provided in the http response in alphabetical order. In the response above,
`namespace` and `subsystem` are placeholders for whatever is configured. If
neither are configured, the metric name would be `example_count`. Usually, we
set the `namespace` to be the project name and `subsystem` as the service name.

## Prometheus Deployment

### Dev

Any manual testing we do while developing is done with a local docker setup. A
few examples of our various docker setups:

- [Glaukos](https://github.com/xmidt-org/glaukos/blob/main/deploy/docker-compose/docker-compose.yml)
- [Svalinn](https://github.com/xmidt-org/codex-deploy/blob/main/deploy/docker-compose/docker-compose.yml)
- [XMiDT](https://github.com/xmidt-org/xmidt/blob/master/deploy/docker-compose/docker-compose.yml)

A prometheus instance is included in the docker setup and can be reached at
`http://localhost:9090/`. All services involved have exposed their metric port,
and Prometheus is
[configured](https://github.com/xmidt-org/glaukos/blob/main/deploy/docker-compose/docFiles/prometheus.yml)
to scrape them. Then, while testing, the developer has visibility into metric
data in addition to output.

### Prod

For production, prometheus is deployed and configured to find services to scrape using [consul](https://prometheus.io/docs/prometheus/latest/configuration/configuration/#consul_sd_config).

## Where to Add Metrics

Generally, we add metrics to track these types of things:
- **HTTP requests**: both [incoming
  (server)](https://pkg.go.dev/github.com/xmidt-org/touchstone@v0.0.3/touchhttp#ServerBundle)
  and [outgoing
  (client)](https://pkg.go.dev/github.com/xmidt-org/touchstone@v0.0.3/touchhttp#ClientBundle);
  tracking the size and duration of a request as well as the response code. Our
  services share some general metrics but often implement their own based on
  what is being done - a common one is counting retries for client requests.
- **Auth**: tracking the
  [result](https://github.com/xmidt-org/bascule/blob/c011b128d6b95fa8358228535c63d1945347adaa/basculehttp/metrics.go#L49)
  of auth middleware on http requests, including
  [why](https://github.com/xmidt-org/bascule/blob/c011b128d6b95fa8358228535c63d1945347adaa/basculehttp/errorResponseReason.go#L24-L34)
  a request was considered unauthorized. With capability checking, we have even
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
- **Complex logic**: this varies by what is trying to be accomplished. Some
  common motivations for adding metrics:
  - if an object being used has a type or enum, there may be a counter with the
    type label, showing the breakdown of what types are being worked with.
  - if there is something long lived (ie connections), a counter of that and any
    easily-grouped metadata is helpful.

### A Few Tips
  - Generally, greater insight into failures is common. Successes are great and
    are counted, but usually we want the most data into things that go wrong.
  - [Labels](https://prometheus.io/docs/practices/naming/#labels) add
    cardinality to metrics. The cardinality of a metric is the cardinality
    (number of possible values) for each label of that metric multiplied
    together. The reason or "why" of a failure in the context of metrics is more
    like a category of failure rather than something as detailed as an error
    message. Enums don't have to be the only label values, but anything that can
    grow without bounds should be considered carefully before being added as a
    metric label value.
