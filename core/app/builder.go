package app

import (
	"github.com/berachain/offchain-sdk/baseapp"
	"github.com/berachain/offchain-sdk/job"
	"github.com/berachain/offchain-sdk/log"
	"github.com/berachain/offchain-sdk/server"

	"github.com/ethereum/go-ethereum/ethdb"
)

// Builder is a builder for an app. It follows a basic factory pattern.
type Builder interface {
	AppName() string
	BuildApp(log.Logger) *baseapp.BaseApp
	RegisterJob(job.Basic)
	RegisterDB(db ethdb.KeyValueStore)
	RegisterHTTPHandler(handler *server.Handler) error
	RegisterPrometheusTelemetry() error
}
