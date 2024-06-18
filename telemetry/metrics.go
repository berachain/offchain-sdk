package telemetry

import (
	"errors"
	"time"

	"github.com/berachain/offchain-sdk/telemetry/datadog"
	"github.com/berachain/offchain-sdk/telemetry/prometheus"
)

// Config serves as a global telemetry configuration.
// Provide the required config for the desired telemetry backend(s).
type Config struct {
	HealthReportInterval time.Duration

	Datadog    datadog.Config
	Prometheus prometheus.Config
}

// NewMetrics returns a new Metrics instance based on the given configuration.
// Enables only the telemetry backend(s) which is/are specified as enabled.
func NewMetrics(cfg *Config) (Metrics, error) {
	var (
		m   = &metrics{}
		err error
	)
	if cfg.Datadog.Enabled {
		if m.datadog, err = datadog.NewMetrics(&cfg.Datadog); err != nil {
			return nil, err
		}
	}
	if cfg.Prometheus.Enabled {
		if m.prometheus, err = prometheus.NewMetrics(&cfg.Prometheus); err != nil {
			return nil, err
		}
	}
	return m, nil
}

type metrics struct {
	datadog    Metrics
	prometheus Metrics
}

func (m *metrics) Gauge(name string, value float64, rate float64, tags ...string) {
	if m.datadog != nil {
		m.datadog.Gauge(name, value, rate, tags...)
	}
	if m.prometheus != nil {
		m.prometheus.Gauge(name, value, rate, tags...)
	}
}

func (m *metrics) Incr(name string, tags ...string) {
	if m.datadog != nil {
		m.datadog.Incr(name, tags...)
	}
	if m.prometheus != nil {
		m.prometheus.Incr(name, tags...)
	}
}

func (m *metrics) Decr(name string, tags ...string) {
	if m.datadog != nil {
		m.datadog.Decr(name, tags...)
	}
	if m.prometheus != nil {
		m.prometheus.Decr(name, tags...)
	}
}

func (m *metrics) Count(name string, value int64, tags ...string) {
	if m.datadog != nil {
		m.datadog.Count(name, value, tags...)
	}
	if m.prometheus != nil {
		m.prometheus.Count(name, value, tags...)
	}
}

func (m *metrics) IncMonotonic(name string, tags ...string) {
	if m.datadog != nil {
		m.datadog.IncMonotonic(name, tags...)
	}
	if m.prometheus != nil {
		m.prometheus.IncMonotonic(name, tags...)
	}
}

func (m *metrics) Error(errName string) {
	if m.datadog != nil {
		m.datadog.Error(errName)
	}
	if m.prometheus != nil {
		m.prometheus.Error(errName)
	}
}

func (m *metrics) Histogram(name string, value float64, rate float64, tags ...string) {
	if m.datadog != nil {
		m.datadog.Histogram(name, value, rate, tags...)
	}
	if m.prometheus != nil {
		m.prometheus.Histogram(name, value, rate, tags...)
	}
}

func (m *metrics) Time(name string, value time.Duration, tags ...string) {
	if m.datadog != nil {
		m.datadog.Time(name, value, tags...)
	}
	if m.prometheus != nil {
		m.prometheus.Time(name, value, tags...)
	}
}

func (m *metrics) Latency(jobName string, start time.Time, tags ...string) {
	if m.datadog != nil {
		m.datadog.Latency(jobName, start, tags...)
	}
	if m.prometheus != nil {
		m.prometheus.Latency(jobName, start, tags...)
	}
}

func (m *metrics) Close() error {
	var err error
	if m.datadog != nil {
		err = m.datadog.Close()
	}
	if m.prometheus != nil {
		err = errors.Join(err, m.prometheus.Close())
	}
	return err
}
