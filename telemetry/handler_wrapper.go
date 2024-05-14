package telemetry

import (
	"context"
	"fmt"
	"time"

	"github.com/berachain/offchain-sdk/log"
	"go-micro.dev/v4/server"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Metrics Wrapper can wrap a HTTP server with Metrics instance
// to collect telemetry for every request/response.
func MetricsWrapper(m Metrics, log log.Logger) server.HandlerWrapper {
	return func(next server.HandlerFunc) server.HandlerFunc {
		return func(c context.Context, req server.Request, rsp interface{}) error {
			metricsTags := getRequestTags(req)

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

func getRequestTags(req server.Request) []string {
	return []string{
		fmt.Sprintf("endpoint:%s", req.Endpoint()), fmt.Sprintf("method:%s", req.Method()),
	}
}
