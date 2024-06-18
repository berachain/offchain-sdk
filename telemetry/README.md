# telemetry

The metrics utility for offchain-sdk.

`./type.go` defines the interface for the supported metrics methods.

By specifying the config, the metrics can be emitted via Datadog and/or Prometheus. Please see
following subsections for detailed configurations

## Datadog

The first step is adding a section in your config file. See following subsection for details. The
source code defining those configs can be found in `./datadog/config.go`

### Datadog Configs

`Enabled`: change this to `true` if the metrics should be emitted to Datadog

`StatsdAddr`: this is the address of Datadog Statsd client. This is needed if the metrics should be
emitted from Datadog

`Namespace`: this will appear as the tag `Namespace` in Datadog

### Datadog Methods

`./datadog/metrics.go` impelements the Datadog version of the supported metrics methods defined in
`./type.go`. All implementation are just simple wrapper, wrapping the native methods in Datadog

## Prometheus

The first step is adding a section in your config file. See following subsection for details. The
source code defining those configs can be found in `./prometheus/config.go`

### Prometheus Configs

`Enabled`: change this to `true` if the metrics should be emitted to Prometheus

`Namespace` and `Subsystem`: those 2 fields will be added as prefix of metrics name. For example
if `Namespace` is `app` and `Subsystem` is `api`, then the full metrics name of `request_success`
will be `app_api_request_success`

`HistogramBucketCount`: This is the count of buckets used for Histogram typed metrics. It is defaulted to 10.

### Prometheus Methods

Different from Datadog, Prometheus only provides
[4 basic metrics type](https://prometheus.io/docs/concepts/metric_types/). As a result,
`./prometheus/metrics.go` implements the metrics methods defined in `./type.go` using the 4 basic
Prometheus metrics. Following subsection documents the methods that with implementation notes. For
more information of the 4 basic Prometheus metrics, please see
[here](https://prometheus.io/docs/tutorials/understanding_metric_types/).

`Gauge`: this method wraps the `Gauge` metrics of Prometheus

`Decr` and `Incr`: those 2 methods are implemented using the `Gauge` metrics of Prometheus

`Count`: this method wraps the `Count` metrics of Prometheus. Please note that after deployment
or restart of instance, `Count` will be reset to 0. This is a by-design feature of Prometheus

`IncMonotonic` and `Error`: those 2 methods are implemented using the `Count` metrics of Prometheus

`Histogram`: this method wraps the `Histogram` metrics of Prometheus with linear bucket

`Time` and `Latency`: those 2 methods are implemented using the `Summary` metrics of Prometheus,
using a pre-defined observation of quantile: p50, p90 and p99
