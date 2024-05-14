package prometheus

import (
	"fmt"
	"strings"
	"time"
	"unicode"

	"github.com/prometheus/client_golang/prometheus"
)

const (
	initialVecCapacity = 32
	tagSlices          = 2
)

type metrics struct {
	cfg *Config

	gaugeVecs     map[string]*prometheus.GaugeVec
	counterVecs   map[string]*prometheus.CounterVec
	histogramVecs map[string]*prometheus.HistogramVec
}

// NewMetrics initializes a new instance of Prometheus metrics.
func NewMetrics(cfg *Config) (*metrics, error) { //nolint:revive // only used as Metrics interface.
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	p := &metrics{cfg: cfg}
	if !cfg.Enabled {
		return p, nil
	}

	p.gaugeVecs = make(map[string]*prometheus.GaugeVec, initialVecCapacity)
	p.counterVecs = make(map[string]*prometheus.CounterVec, initialVecCapacity)
	p.histogramVecs = make(map[string]*prometheus.HistogramVec, initialVecCapacity)
	return p, nil
}

func (p *metrics) Close() error {
	return nil
}

// Gauge implements the Gauge method of the Metrics interface using GaugeVec.
func (p *metrics) Gauge(name string, value float64, tags []string, _ float64) {
	if !p.cfg.Enabled {
		return
	}

	name = forceValidName(name)
	labels, labelValues := parseTagsToLabelPairs(tags)
	gaugeVec, exists := p.gaugeVecs[name]
	if !exists {
		gaugeVec = prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name:      name,
			Namespace: p.cfg.Namespace,
			Subsystem: p.cfg.Subsystem,
			Help:      name + " gauge",
		}, labels)
		prometheus.MustRegister(gaugeVec)
		p.gaugeVecs[name] = gaugeVec
	}
	gaugeVec.WithLabelValues(labelValues...).Set(value)
}

// Incr implements the Incr method of the Metrics interface using GaugeVec.
func (p *metrics) Incr(name string, tags []string) {
	if !p.cfg.Enabled {
		return
	}

	name = forceValidName(name)
	labels, labelValues := parseTagsToLabelPairs(tags)
	gaugeVec, exists := p.gaugeVecs[name]
	if !exists {
		gaugeVec = prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name:      name,
			Namespace: p.cfg.Namespace,
			Subsystem: p.cfg.Subsystem,
			Help:      name + " incr/decr gauge",
		}, labels)
		prometheus.MustRegister(gaugeVec)
		p.gaugeVecs[name] = gaugeVec
	}
	gaugeVec.WithLabelValues(labelValues...).Inc()
}

// Decr implements the Decr method of the Metrics interface using GaugeVec.
func (p *metrics) Decr(name string, tags []string) {
	if !p.cfg.Enabled {
		return
	}

	name = forceValidName(name)
	labels, labelValues := parseTagsToLabelPairs(tags)
	gaugeVec, exists := p.gaugeVecs[name]
	if !exists {
		gaugeVec = prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name:      name,
			Namespace: p.cfg.Namespace,
			Subsystem: p.cfg.Subsystem,
			Help:      name + " incr/decr gauge",
		}, labels)
		prometheus.MustRegister(gaugeVec)
		p.gaugeVecs[name] = gaugeVec
	}
	gaugeVec.WithLabelValues(labelValues...).Dec()
}

// Count implements the Count method of the Metrics interface using CounterVec.
func (p *metrics) Count(name string, value int64, tags []string) {
	if !p.cfg.Enabled {
		return
	}

	name = forceValidName(name)
	labels, labelValues := parseTagsToLabelPairs(tags)
	counterVec, exists := p.counterVecs[name]
	if !exists {
		counterVec = prometheus.NewCounterVec(prometheus.CounterOpts{
			Name:      name,
			Namespace: p.cfg.Namespace,
			Subsystem: p.cfg.Subsystem,
			Help:      name + " counter",
		}, labels)
		prometheus.MustRegister(counterVec)
		p.counterVecs[name] = counterVec
	}
	counterVec.WithLabelValues(labelValues...).Add(float64(value))
}

// IncMonotonic implements the IncMonotonic method of the Metrics interface using CounterVec.
func (p *metrics) IncMonotonic(name string, tags []string) {
	if !p.cfg.Enabled {
		return
	}

	name = forceValidName(name)
	labels, labelValues := parseTagsToLabelPairs(tags)
	counterVec, exists := p.counterVecs[name]
	if !exists {
		counterVec = prometheus.NewCounterVec(prometheus.CounterOpts{
			Name:      name,
			Namespace: p.cfg.Namespace,
			Subsystem: p.cfg.Subsystem,
			Help:      name + " counter",
		}, labels)
		prometheus.MustRegister(counterVec)
		p.counterVecs[name] = counterVec
	}
	counterVec.WithLabelValues(labelValues...).Inc()
}

// Error implements the Error method of the Metrics interface using CounterVec.
func (p *metrics) Error(errName string) {
	p.IncMonotonic("stats.errors", []string{fmt.Sprintf("type:%s", errName)})
}

// Histogram implements the Histogram method of the Metrics interface using HistogramVec.
func (p *metrics) Histogram(name string, value float64, tags []string, rate float64) {
	if !p.cfg.Enabled {
		return
	}

	name = forceValidName(name)
	labels, labelValues := parseTagsToLabelPairs(tags)
	histogramVec, exists := p.histogramVecs[name]
	if !exists {
		histogramVec = prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name:      name,
			Namespace: p.cfg.Namespace,
			Subsystem: p.cfg.Subsystem,
			Help:      name + " histogram",
			Buckets:   prometheus.LinearBuckets(0, rate, 10), // Adjust bucketing as necessary
		}, labels)
		prometheus.MustRegister(histogramVec)
		p.histogramVecs[name] = histogramVec
	}
	histogramVec.WithLabelValues(labelValues...).Observe(value)
}

// Time implements the Time method of the Metrics interface using GaugeVec.
func (p *metrics) Time(name string, value time.Duration, tags []string) {
	if !p.cfg.Enabled {
		return
	}

	name = forceValidName(name)
	labels, labelValues := parseTagsToLabelPairs(tags)
	histogramVec, exists := p.histogramVecs[name]
	if !exists {
		histogramVec = prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name:      name,
			Namespace: p.cfg.Namespace,
			Subsystem: p.cfg.Subsystem,
			Help:      name + " timing histogram",
			Buckets:   prometheus.LinearBuckets(0, 1, 10), // Adjust bucketing as necessary
		}, labels)
		prometheus.MustRegister(histogramVec)
		p.histogramVecs[name] = histogramVec
	}

	// Convert time.Duration to seconds since Prometheus prefers base units
	histogramVec.WithLabelValues(labelValues...).Observe(value.Seconds())
}

// Latency is a helper function to measure the latency of a routine.
func (p *metrics) Latency(jobName string, start time.Time, tags ...string) {
	p.Time("stats.latency", time.Since(start), append(tags, fmt.Sprintf("job:%s", jobName)))
}

// parseTagsToLabelPairs converts a slice of tags in "key:value" format to two slices:
// one for the label names and one for the label values, maintaining order.
func parseTagsToLabelPairs(tags []string) ([]string, []string) {
	labels := make([]string, 0, len(tags))
	labelValues := make([]string, 0, len(tags))
	for _, tag := range tags {
		kv := strings.SplitN(tag, ":", tagSlices)
		if len(kv) == tagSlices {
			labels = append(labels, kv[0])
			labelValues = append(labelValues, kv[1])
		}
	}
	return labels, labelValues
}

// forceValidName converts a string to a valid Prometheus metric name.
// ref: https://prometheus.io/docs/concepts/data_model/#metric-names-and-labels
func forceValidName(name string) string {
	// Convert the input string to a slice of runes to properly handle potentially multi-byte
	// characters.
	runes := []rune(name)

	// Process the first character separately to ensure it matches `[a-zA-Z_:]`
	if len(runes) > 0 {
		if !unicode.IsLetter(runes[0]) && runes[0] != '_' && runes[0] != ':' {
			runes[0] = '_'
		}
	}

	// Process the rest of the characters to ensure they match `[a-zA-Z0-9_:]`
	for i := 1; i < len(runes); i++ {
		if !unicode.IsLetter(runes[i]) && !unicode.IsDigit(runes[i]) &&
			runes[i] != '_' && runes[i] != ':' {
			runes[i] = '_'
		}
	}

	return string(runes)
}
