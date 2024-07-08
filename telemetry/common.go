package telemetry

import "fmt"

// ParseLabelPairsToTags converts a list of label names and values to a list of tags,
// in format of "name:value".
func ParseLabelPairsToTags(labels, labelValues []string) []string {
	minLen := min(len(labels), len(labelValues))
	tags := make([]string, 0, minLen)
	for i := 0; i < minLen; i++ {
		tags = append(tags, fmt.Sprintf("%s:%s", labels[i], labelValues[i]))
	}
	return tags
}
