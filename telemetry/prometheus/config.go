package prometheus

import (
	"fmt"
	"regexp"
)

const (
	// Default bucket count 1000 can satisfy the precision of p99 for most histogram stats.
	DefaultBucketCount = 1000
)

type Config struct {
	Enabled              bool
	Namespace            string // optional
	Subsystem            string // optional
	HistogramBucketCount int    // Number of buckets for histogram, default to 1000
	// Number of buckets for time buckets, default to 1000.
	// The bucket size is 10ms, so the maximum covered time range is 10ms * TimeBucketCount.
	TimeBucketCount int
}

func (c *Config) Validate() error {
	if !c.Enabled {
		return nil
	}

	if c.Namespace != "" && !validPromMetric(c.Namespace) {
		return fmt.Errorf("invalid prometheus namespace: %s", c.Namespace)
	}

	if c.Subsystem != "" && !validPromMetric(c.Subsystem) {
		return fmt.Errorf("invalid prometheus subsystem: %s", c.Subsystem)
	}

	return nil
}

// Note that Go's regex engine requires double escaping in string literals, hence `\\` is used
// instead of `\`.
var promMetricRe = regexp.MustCompile(`^[a-zA-Z_:][a-zA-Z0-9_:]*$`)

// validPromMetric checks if the namespace adheres to the pattern `[a-zA-Z_:][a-zA-Z0-9_:]*`
// ref: https://prometheus.io/docs/concepts/data_model/#metric-names-and-labels
func validPromMetric(namespace string) bool {
	// Use the MatchString method to check if the string matches the pattern.
	return promMetricRe.MatchString(namespace)
}
