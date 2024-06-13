package telemetry

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/berachain/offchain-sdk/log"
	"github.com/gogo/status"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"go-micro.dev/v4/server"
	"google.golang.org/grpc/codes"
)

// telemetryRespWriter is a wrapper around http.ResponseWriter that captures the status code.
type telemetryRespWriter struct {
	http.ResponseWriter
	statusCode int
}

func newTelemetryRespWriter(w http.ResponseWriter) *telemetryRespWriter {
	// Default to 200 OK in case the handler does not explicitly set the status code.
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

// WrapHTTPHandler wraps a HTTP server with the given Metrics instance.
// to collect telemetry for every request/response.
func WrapHTTPHandler(m Metrics, log log.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			log.Debug("Request received", "method", r.Method, "path", r.URL.Path)
			customWriter := newTelemetryRespWriter(w)
			metricsTags := getHTTPRequestTags(r)

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

// WrapMicroServerHandler wraps a Micro server with the given Metrics instance
// to collect telemetry for every request/response.
func WrapMicroServerHandler(m Metrics, log log.Logger) server.HandlerWrapper {
	return func(next server.HandlerFunc) server.HandlerFunc {
		return func(c context.Context, req server.Request, rsp interface{}) error {
			metricsTags := getMicroRequestTags(req)

			// Increment request count metric under `request.count`
			m.IncMonotonic("request.count", metricsTags)

			start := time.Now()
			err := next(c, req, rsp)

			// Record latency metric under `response.latency`
			m.Time("request.latency", time.Since(start), metricsTags)

			// Separately record errors under `request.errors`
			if err != nil {
				code := status.Code(err)
				if code == codes.Internal {
					log.Error("Internal error", "error", err, "request", req.Endpoint())
				}

				metricsTags = append(metricsTags, fmt.Sprintf("code:%s", code.String()))
				m.IncMonotonic("request.errors", metricsTags)
			}

			return err
		}
	}
}

func getMicroRequestTags(req server.Request) []string {
	return []string{
		fmt.Sprintf("endpoint:%s", req.Endpoint()), fmt.Sprintf("method:%s", req.Method()),
	}
}

func getHTTPRequestTags(req *http.Request) []string {
	// if the request is a gRPC-gateway request, use the gRPC method as the endpoint
	// i.e. "/package.service/method"
	rpcMethod, ok := runtime.RPCMethod(req.Context())
	if !ok {
		rpcMethod = req.URL.Path
	}
	return []string{
		fmt.Sprintf("endpoint:%s", rpcMethod), fmt.Sprintf("method:%s", req.Method),
	}
}
