package log_test

import (
	"bytes"
	"testing"

	"github.com/berachain/offchain-sdk/log"
)

func TestLogger(t *testing.T) {
	// Create a buffer to capture the log output
	var buf bytes.Buffer

	// Create a new logger with the buffer as the destination
	logger := log.NewLogger(&buf, "test-runner")

	// Log some messages at different levels
	logger.Info("Info message")
	logger.Error("Error message")
	logger.Debug("Debug message")

	// Retrieve the log output from the buffer
	output := buf.String()

	// Assert that the log output contains the expected messages
	if !contains(output, "Info message") {
		t.Errorf("Expected log output to contain 'Info message', got: %s", output)
	}
	if !contains(output, "Error message") {
		t.Errorf("Expected log output to contain 'Error message', got: %s", output)
	}
	if !contains(output, "Debug message") {
		t.Errorf("Expected log output to contain 'Debug message', got: %s", output)
	}
}

func contains(s, substr string) bool {
	return bytes.Contains([]byte(s), []byte(substr))
}
