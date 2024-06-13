package telemetry

import (
	"fmt"
	"net/http"
	"time"

	"github.com/berachain/offchain-sdk/log"
)

type telemetryRespWriter struct {
	http.ResponseWriter
	statusCode int
}

func newTelemetryRespWriter(w http.ResponseWriter) *telemetryRespWriter {
	// Default to 200 OK in case the handler does not explicitly set the status code
	return &telemetryRespWriter{ResponseWriter: w, statusCode: http.StatusOK}
}

func (rw *telemetryRespWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *telemetryRespWriter) Write(b []byte) (int, error) {
	if rw.statusCode == http.StatusOK && len(b) > 0 {
		// When Write is called without WriteHeader, infer the status code
		rw.statusCode = http.StatusOK
	}
	return rw.ResponseWriter.Write(b)
}

// GetHandlerWrapper wraps a HTTP server with the given Metrics instance.
// to collect telemetry for every request/response.
func GetHandlerWrapper(m Metrics, log log.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			customWriter := newTelemetryRespWriter(w)
			metricsTags := getRequestTags(r)

			// Increment request count metric under `request.count`
			m.IncMonotonic("request.count", metricsTags)

			start := time.Now()
			next.ServeHTTP(customWriter, r)

			// Record latency metric under `response.latency`
			m.Time("request.latency", time.Since(start), metricsTags)

			// Separately record errors under `request.errors`
			if customWriter.statusCode >= http.StatusBadRequest {
				if customWriter.statusCode >= http.StatusInternalServerError {
					log.Error(
						"Internal error",
						"statusCode", customWriter.statusCode,
						"method", r.Method,
						"path", r.URL.Path,
					)
				}

				metricsTags = append(metricsTags, fmt.Sprintf("code:%d", customWriter.statusCode))
				m.IncMonotonic("request.errors", metricsTags)
			}
		})
	}
}

func getRequestTags(req *http.Request) []string {
	return []string{
		fmt.Sprintf("endpoint:%s", req.URL.Path), fmt.Sprintf("method:%s", req.Method),
	}
}
