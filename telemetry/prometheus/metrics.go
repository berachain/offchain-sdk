package prometheus

import (
	"errors"
	"fmt"
	"strings"
	"time"
	"unicode"

	"github.com/berachain/offchain-sdk/tools/rwstore"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	initialVecCapacity = 32
	tagSlices          = 2

	// Constant for Summary (quantile) metrics.
	quantile50 = 0.5
	quantile90 = 0.9
	quantile99 = 0.99
	// Usually the allowed relative error margin is 10%, and the absolute error margin is computed
	// using (1 - quantile) * 10%. For example for p90, error margin is (1 - 90%) * 10% = 1% = 0.01.
	errorMargin50 = 0.05
	errorMargin90 = 0.01
	errorMargin99 = 0.001
)

type metrics struct {
	cfg *Config

	gaugeVecs     *rwstore.RWMap[string, *prometheus.GaugeVec]
	counterVecs   *rwstore.RWMap[string, *prometheus.CounterVec]
	histogramVecs *rwstore.RWMap[string, *prometheus.HistogramVec]
	summaryVecs   *rwstore.RWMap[string, *prometheus.SummaryVec]
}

// NewMetrics initializes a new instance of Prometheus metrics.
func NewMetrics(cfg *Config) (*metrics, error) { //nolint:revive // only used as Metrics interface.
	setDefaultCfg(cfg)
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	p := &metrics{cfg: cfg}
	if !cfg.Enabled {
		return p, nil
	}

	p.gaugeVecs = rwstore.NewRWMap[string, *prometheus.GaugeVec]()
	p.counterVecs = rwstore.NewRWMap[string, *prometheus.CounterVec]()
	p.histogramVecs = rwstore.NewRWMap[string, *prometheus.HistogramVec]()
	p.summaryVecs = rwstore.NewRWMap[string, *prometheus.SummaryVec]()
	return p, nil
}

func (p *metrics) Close() error {
	return nil
}

// Gauge implements the Gauge method of the Metrics interface using GaugeVec.
func (p *metrics) Gauge(name string, value float64, _ float64, tags ...string) {
	if !p.cfg.Enabled {
		return
	}

	name = forceValidName(name)
	labels, labelValues := parseTagsToLabelPairs(tags)

	if gaugeVec, exists := p.gaugeVecs.Get(name); exists {
		gaugeVec.WithLabelValues(labelValues...).Set(value)
		return
	}

	gaugeVec := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name:      name,
		Namespace: p.cfg.Namespace,
		Subsystem: p.cfg.Subsystem,
		Help:      name + " gauge",
	}, labels)

	if err := prometheus.Register(gaugeVec); err != nil {
		// In case of concurrent registration, get the one that has already registered
		var alreadyRegisteredError prometheus.AlreadyRegisteredError
		if errors.As(err, &alreadyRegisteredError) {
			//nolint:errcheck // OK
			gaugeVec = alreadyRegisteredError.ExistingCollector.(*prometheus.GaugeVec)
		} else {
			// Otherwise we should panic to fail fast
			panic(err)
		}
	} else {
		p.gaugeVecs.Set(name, gaugeVec)
	}

	gaugeVec.WithLabelValues(labelValues...).Set(value)
}

// Incr implements the Incr method of the Metrics interface using GaugeVec.
func (p *metrics) Incr(name string, tags ...string) {
	if !p.cfg.Enabled {
		return
	}

	name = forceValidName(name)
	labels, labelValues := parseTagsToLabelPairs(tags)

	if gaugeVec, exists := p.gaugeVecs.Get(name); exists {
		gaugeVec.WithLabelValues(labelValues...).Inc()
		return
	}

	gaugeVec := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name:      name,
		Namespace: p.cfg.Namespace,
		Subsystem: p.cfg.Subsystem,
		Help:      name + " incr/decr gauge",
	}, labels)

	if err := prometheus.Register(gaugeVec); err != nil {
		// In case of concurrent registration, get the one that has already registered
		var alreadyRegisteredError prometheus.AlreadyRegisteredError
		if errors.As(err, &alreadyRegisteredError) {
			//nolint:errcheck // OK
			gaugeVec = alreadyRegisteredError.ExistingCollector.(*prometheus.GaugeVec)
		} else {
			// Otherwise we should panic to fail fast
			panic(err)
		}
	} else {
		p.gaugeVecs.Set(name, gaugeVec)
	}

	gaugeVec.WithLabelValues(labelValues...).Inc()
}

// Decr implements the Decr method of the Metrics interface using GaugeVec.
func (p *metrics) Decr(name string, tags ...string) {
	if !p.cfg.Enabled {
		return
	}

	name = forceValidName(name)
	labels, labelValues := parseTagsToLabelPairs(tags)

	if gaugeVec, exists := p.gaugeVecs.Get(name); exists {
		gaugeVec.WithLabelValues(labelValues...).Dec()
		return
	}

	gaugeVec := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name:      name,
		Namespace: p.cfg.Namespace,
		Subsystem: p.cfg.Subsystem,
		Help:      name + " incr/decr gauge",
	}, labels)

	if err := prometheus.Register(gaugeVec); err != nil {
		// In case of concurrent registration, get the one that has already registered
		var alreadyRegisteredError prometheus.AlreadyRegisteredError
		if errors.As(err, &alreadyRegisteredError) {
			//nolint:errcheck // OK
			gaugeVec = alreadyRegisteredError.ExistingCollector.(*prometheus.GaugeVec)
		} else {
			// Otherwise we should panic to fail fast
			panic(err)
		}
	} else {
		p.gaugeVecs.Set(name, gaugeVec)
	}
	gaugeVec.WithLabelValues(labelValues...).Dec()
}

// Count implements the Count method of the Metrics interface using CounterVec.
func (p *metrics) Count(name string, value int64, tags ...string) {
	if !p.cfg.Enabled {
		return
	}

	name = forceValidName(name)
	labels, labelValues := parseTagsToLabelPairs(tags)

	if counterVec, exists := p.counterVecs.Get(name); exists {
		counterVec.WithLabelValues(labels...).Add(float64(value))
		return
	}

	counterVec := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name:      name,
		Namespace: p.cfg.Namespace,
		Subsystem: p.cfg.Subsystem,
		Help:      name + " counter",
	}, labels)
	if err := prometheus.Register(counterVec); err != nil {
		// In case of concurrent registration, get the one that has already registered
		var alreadyRegisteredError prometheus.AlreadyRegisteredError
		if errors.As(err, &alreadyRegisteredError) {
			//nolint:errcheck // OK
			counterVec = alreadyRegisteredError.ExistingCollector.(*prometheus.CounterVec)
		} else {
			// Otherwise we should panic to fail fast
			panic(err)
		}
	} else {
		p.counterVecs.Set(name, counterVec)
	}

	counterVec.WithLabelValues(labelValues...).Add(float64(value))
}

// IncMonotonic implements the IncMonotonic method of the Metrics interface using CounterVec.
func (p *metrics) IncMonotonic(name string, tags ...string) {
	if !p.cfg.Enabled {
		return
	}

	name = forceValidName(name)
	labels, labelValues := parseTagsToLabelPairs(tags)

	if counterVec, exists := p.counterVecs.Get(name); exists {
		counterVec.WithLabelValues(labels...).Inc()
		return
	}

	counterVec := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name:      name,
		Namespace: p.cfg.Namespace,
		Subsystem: p.cfg.Subsystem,
		Help:      name + " counter",
	}, labels)
	if err := prometheus.Register(counterVec); err != nil {
		// In case of concurrent registration, get the one that has already registered
		var alreadyRegisteredError prometheus.AlreadyRegisteredError
		if errors.As(err, &alreadyRegisteredError) {
			//nolint:errcheck // OK
			counterVec = alreadyRegisteredError.ExistingCollector.(*prometheus.CounterVec)
		} else {
			// Otherwise we should panic to fail fast
			panic(err)
		}
	} else {
		p.counterVecs.Set(name, counterVec)
	}

	counterVec.WithLabelValues(labelValues...).Inc()
}

// Error implements the Error method of the Metrics interface using CounterVec.
func (p *metrics) Error(errName string) {
	p.IncMonotonic("errors", fmt.Sprintf("type:%s", errName))
}

// Histogram implements the Histogram method of the Metrics interface using HistogramVec with
// linear buckets.
// TODO: Support different types of buckets beyond linear buckets in future implementations.
func (p *metrics) Histogram(name string, value float64, rate float64, tags ...string) {
	if !p.cfg.Enabled {
		return
	}

	name = forceValidName(name)
	labels, labelValues := parseTagsToLabelPairs(tags)

	if histogramVec, exists := p.histogramVecs.Get(name); exists {
		histogramVec.WithLabelValues(labelValues...).Observe(value)
		return
	}

	histogramVec := prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:      name,
		Namespace: p.cfg.Namespace,
		Subsystem: p.cfg.Subsystem,
		Help:      name + " histogram",
		// The maximum covered stats range is rate * HistogramBucketCount
		Buckets: prometheus.LinearBuckets(0, rate, p.cfg.HistogramBucketCount),
	}, labels)

	if err := prometheus.Register(histogramVec); err != nil {
		// In case of concurrent registration, get the one that has already registered
		var alreadyRegisteredError prometheus.AlreadyRegisteredError
		if errors.As(err, &alreadyRegisteredError) {
			//nolint:errcheck // OK
			histogramVec = alreadyRegisteredError.ExistingCollector.(*prometheus.HistogramVec)
		} else {
			// Otherwise we should panic to fail fast
			panic(err)
		}
	} else {
		p.histogramVecs.Set(name, histogramVec)
	}

	histogramVec.WithLabelValues(labelValues...).Observe(value)
}

// Time implements the Time method of the Metrics interface using SummaryVec.
// Currently the p50/p90/p99 quantiles are recorded.
func (p *metrics) Time(name string, value time.Duration, tags ...string) {
	if !p.cfg.Enabled {
		return
	}

	name = forceValidName(name)
	labels, labelValues := parseTagsToLabelPairs(tags)
	summaryVec, ok := p.summaryVecs.Get(name)
	if ok {
		summaryVec.WithLabelValues(labels...).Observe(value.Seconds())
		return
	}

	summaryVec = prometheus.NewSummaryVec(prometheus.SummaryOpts{
		Name:      name,
		Namespace: p.cfg.Namespace,
		Subsystem: p.cfg.Subsystem,
		Help:      name + " timing summary",
		Objectives: map[float64]float64{
			quantile50: errorMargin50,
			quantile90: errorMargin90,
			quantile99: errorMargin99,
		},
	}, labels)
	if err := prometheus.Register(summaryVec); err != nil {
		// In case of concurrent registration, get the one that has already registered
		var alreadyRegisteredError prometheus.AlreadyRegisteredError
		if errors.As(err, &alreadyRegisteredError) {
			//nolint:errcheck // OK
			summaryVec = alreadyRegisteredError.ExistingCollector.(*prometheus.SummaryVec)
		} else {
			// Otherwise we should panic to fail fast
			panic(err)
		}
	} else {
		p.summaryVecs.Set(name, summaryVec)
	}
	// Convert time.Duration to seconds since Prometheus prefers base units
	// see https://prometheus.io/docs/practices/naming/#base-units
	summaryVec.WithLabelValues(labelValues...).Observe(value.Seconds())
}

// Latency is a helper function to measure the latency of a routine.
func (p *metrics) Latency(jobName string, start time.Time, tags ...string) {
	p.Time(fmt.Sprintf("%s.latency", jobName), time.Since(start), tags...)
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

// Set default values if not provided.
func setDefaultCfg(cfg *Config) {
	if cfg.HistogramBucketCount <= 0 {
		cfg.HistogramBucketCount = DefaultBucketCount
	}
}
