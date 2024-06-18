package prometheus

import (
	"fmt"
	"regexp"
)

const (
	// Default bucket count for histogram metrics.
	DefaultBucketCount = 10
)

type Config struct {
	Enabled              bool
	Namespace            string // optional
	Subsystem            string // optional
	HistogramBucketCount int    // Number of linear buckets for histogram, default to 10
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
