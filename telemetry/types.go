package telemetry

import "time"

// Metrics are used to record telemetry data. The primary types of telemetry data supported are
// Gauges, Counters, and Histograms; the rest are convenience wrappers around these.
type Metrics interface {
	Gauge(name string, value float64, tags []string, rate float64)

	Incr(name string, tags []string) // should be used if the metric can go up or down
	Decr(name string, tags []string) // should be used if the metric can go up or down

	Count(name string, value int64, tags []string)

	IncMonotonic(name string, tags []string) // should be used if the metric can go up only

	Error(errName string)

	Histogram(name string, value float64, tags []string, rate float64)

	Time(name string, value time.Duration, tags []string)
	Latency(jobName string, start time.Time, tags ...string)

	Close() error
}
