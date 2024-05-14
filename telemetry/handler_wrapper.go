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

			start := time.Now()
			err := next(c, req, rsp)
			m.Time("request.latency", time.Since(start), metricsTags)

			if err != nil {
				code := status.Code(err)
				if code == codes.Internal {
					log.Error("Internal error", "error", err, "request", req.Endpoint())
				}
				tags := append(metricsTags, fmt.Sprintf("code:%s", code.String()))
				m.IncMonotonic("request.errors", tags)
			}

			m.IncMonotonic("request.count", metricsTags)
			return err
		}
	}
}

func getRequestTags(req server.Request) []string {
	return []string{
		fmt.Sprintf("endpoint:%s", req.Endpoint()), fmt.Sprintf("method:%s", req.Method()),
	}
}
