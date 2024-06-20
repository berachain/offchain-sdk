package app

import (
	"github.com/berachain/offchain-sdk/baseapp"
	"github.com/berachain/offchain-sdk/job"
	"github.com/berachain/offchain-sdk/log"
	"github.com/berachain/offchain-sdk/server"
	"github.com/berachain/offchain-sdk/telemetry"

	"github.com/ethereum/go-ethereum/ethdb"
)

// Builder is a builder for an app. It follows a basic factory pattern.
type Builder interface {
	AppName() string
	BuildApp(log.Logger) *baseapp.BaseApp
	RegisterJob(job.Basic)
	RegisterMetrics(cfg *telemetry.Config) error
	RegisterDB(db ethdb.KeyValueStore)
	RegisterHTTPHandler(handler *server.Handler) error
	RegisterMiddleware(m server.Middleware) error
	RegisterPrometheusTelemetry() error
}
