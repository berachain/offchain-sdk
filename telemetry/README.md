# Telemetry

The metrics utility for offchain-sdk.

[types.go](./types.go) defines the interface for the supported metrics methods.

By specifying the configuration, the metrics can be emitted via Datadog and/or Prometheus.
Please see the following subsections for detailed configurations.

## Datadog

### Configuration

The first step is adding a section in your config file. See following subsection for details. The
source code defining those configs can be found in [config.go](./datadog/config.go).

#### Datadog Configs

* `Enabled`: Set to `true` to enable metrics emission to Datadog.

* `StatsdAddr`: The address of the Datadog StatsD client. This is needed if the metrics should be
emitted from Datadog.

* `Namespace`: This will appear as the `Namespace` tag in Datadog.

### Datadog Methods

[metrics.go](./datadog/metrics.go) implements the Datadog version of the supported metrics methods
defined in [types.go](./types.go). All implementations are simple wrappers around the native methods
provided by the Datadog `statsd` client.

## Prometheus

### Configuration

The first step is to add a section in your config file. The source code defining these configs can
be found in  [config.go](./prometheus/config.go).

#### Prometheus Configs

* `Enabled`: Set to true to enable metrics emission to Prometheus.

* `Namespace` and `Subsystem`: These fields will be added as prefixes to the metrics name.
For example, if `Namespace` is `app` and `Subsystem` is `api`, then the full metrics name of
`request_success` will be `app_api_request_success`.

* `HistogramBucketCount`: The number of buckets used for Histogram typed metrics. Default is 10.

### Prometheus Methods

Different from Datadog, Prometheus only provides
[4 basic metrics type](https://prometheus.io/docs/concepts/metric_types/). As a result,
[metrics.go](./prometheus/metrics.go) implements the metrics methods defined in [type.go](./type.go)
using these four basic Prometheus metrics. The following subsection documents the methods with
implementation notes. For more information on the four basic Prometheus metrics, please see
[here](https://prometheus.io/docs/tutorials/understanding_metric_types/).

* `Gauge`: This method wraps the `Gauge` metrics of Prometheus.

* `Decr` and `Incr`: Implemented using the `Gauge` metrics of Prometheus.

* `Count`: This method wraps the `Count` metrics of Prometheus. Note that after deployment or instance
restart, `Count` will reset to 0. This is by design in Prometheus.

* `IncMonotonic` and `Error`: Implemented using the `Count` metrics of Prometheus.

* `Histogram`: This method wraps the `Histogram` metrics of Prometheus with linear buckets.

* `Time` and `Latency`: Implemented using the `Summary` metrics of Prometheus, with pre-defined
quantile observations: p50, p90, and p99.
