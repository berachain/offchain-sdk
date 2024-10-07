package app

import (
	"github.com/berachain/offchain-sdk/v2/baseapp"
	"github.com/berachain/offchain-sdk/v2/job"
	"github.com/berachain/offchain-sdk/v2/server"
	"github.com/berachain/offchain-sdk/v2/telemetry"

	"github.com/ethereum/go-ethereum/ethdb"
)

// Builder is a builder for an app. It follows a basic factory pattern.
type Builder interface {
	AppName() string
	BuildApp() *baseapp.BaseApp
	RegisterJob(job.Basic)
	RegisterMetrics(cfg *telemetry.Config) error
	RegisterDB(db ethdb.KeyValueStore)
	RegisterHTTPHandler(handler *server.Handler) error
	RegisterMiddleware(m server.Middleware) error
}
