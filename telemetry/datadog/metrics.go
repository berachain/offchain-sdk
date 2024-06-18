package datadog

import (
	"fmt"
	"time"

	"github.com/DataDog/datadog-go/v5/statsd"
)

type metrics struct {
	enabled bool
	client  *statsd.Client
}

func NewMetrics(cfg *Config) (*metrics, error) { //nolint:revive // only used as Metrics interface.
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	m := &metrics{enabled: cfg.Enabled}
	if !m.enabled {
		return m, nil
	}

	client, err := statsd.New(cfg.StatsdAddr, statsd.WithNamespace(cfg.Namespace))
	if err != nil {
		return nil, fmt.Errorf("failed to create statsd client for %s: %w", cfg.StatsdAddr, err)
	}
	m.client = client

	return m, nil
}

func (m *metrics) Close() error {
	if !m.enabled {
		return nil
	}
	return m.client.Close()
}

func (m *metrics) Gauge(name string, value float64, rate float64, tags ...string) {
	if !m.enabled {
		return
	}
	//#nosec:G104 // handled by m.client.Gauge()
	m.client.Gauge(name, value, tags, rate) //nolint:errcheck // handled by m.client.Gauge()
}

func (m *metrics) Count(name string, value int64, tags ...string) {
	if !m.enabled {
		return
	}
	//#nosec:G104 // handled by m.client.Count()
	m.client.Count(name, value, tags, 1) //nolint:errcheck // handled by m.client.Count()
}

func (m *metrics) IncMonotonic(name string, tags ...string) {
	m.Incr(name, tags...)
}

func (m *metrics) Incr(name string, tags ...string) {
	if !m.enabled {
		return
	}
	//#nosec:G104 // handled by m.client.Incr()
	m.client.Incr(name, tags, 1) //nolint:errcheck // handled by m.client.Incr()
}

func (m *metrics) Decr(name string, tags ...string) {
	if !m.enabled {
		return
	}
	//#nosec:G104 // handled by m.client.Decr()
	m.client.Decr(name, tags, 1) //nolint:errcheck // handled by m.client.Decr()
}

func (m *metrics) Set(name string, value string, tags ...string) {
	if !m.enabled {
		return
	}
	//#nosec:G104 // handled by m.client.Set()
	m.client.Set(name, value, tags, 1) //nolint:errcheck // handled by m.client.Set()
}

func (m *metrics) Histogram(name string, value float64, rate float64, tags ...string) {
	if !m.enabled {
		return
	}
	//#nosec:G104 // handled by m.client.Histogram()
	m.client.Histogram(name, value, tags, rate) //nolint:errcheck // handled by m.client.Histogram()
}

func (m *metrics) Time(name string, value time.Duration, tags ...string) {
	if !m.enabled {
		return
	}
	//#nosec:G104 // handled by m.client.Timing()
	m.client.Timing(name, value, tags, 1) //nolint:errcheck // handled by m.client.Timing()
}

func (m *metrics) Error(errName string) {
	m.Incr("stats.errors", fmt.Sprintf("type:%s", errName))
}

// Latency is a helper function to measure the latency of a routine.
func (m *metrics) Latency(jobName string, start time.Time, tags ...string) {
	m.Time("stats.latency", time.Since(start), append(tags, fmt.Sprintf("job:%s", jobName))...)
}
