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

`./datadog/metrics.go` impelements the Datadog version of the supported metrics method defined in
`./type.go`. All implementation are just simple wrapper, wrapping the native methods in Datadog

## Prometheus

The first step is adding a section in your config file. See following subsection for details. The
source code defining those configs can be found in `./prometheus/config.go`
`Enabled`: change this to `true` if the metrics should be emitted to Prometheus
`Namespace` and `Subsystem`: those 2 fields will be added as prefix of metrics name. For example
if `Namespace` is `app` and `Subsystem` is `api`, then the full metrics name of `request_success`
will be `app_api_request_success`
`HistogramBucketCount`: This is the count of buckets used for Histogram typed metrics. It is defaulted to 10.
