package baseapp

import (
	"context"

	"github.com/berachain/offchain-sdk/v2/client/eth"
	"github.com/berachain/offchain-sdk/v2/job"
	"github.com/berachain/offchain-sdk/v2/log"
	"github.com/berachain/offchain-sdk/v2/server"
	"github.com/berachain/offchain-sdk/v2/telemetry"

	ethdb "github.com/ethereum/go-ethereum/ethdb"
)

// BaseApp is the base application.
type BaseApp struct {
	// name is the name of the application
	name string

	// logger is the logger for the baseapp.
	logger log.Logger

	// jobMgr
	jobMgr *JobManager

	// svr is the server for the baseapp.
	svr *server.Server

	// chain is the eth client for the baseapp.
	chain eth.Client
}

// New creates a new baseapp.
func New(
	name string,
	logger log.Logger,
	ethClient eth.Client,
	jobs []job.Basic,
	db ethdb.KeyValueStore,
	svr *server.Server,
	metrics telemetry.Metrics,
) *BaseApp {
	return &BaseApp{
		name:   name,
		logger: logger,
		jobMgr: NewManager(
			jobs,
			&contextFactory{
				connPool: ethClient,
				logger:   logger,
				db:       db,
				metrics:  metrics,
			},
		),
		svr: svr,
	}
}

// Logger returns the logger for the baseapp.
func (b *BaseApp) Logger() log.Logger {
	return b.logger.With("namespace", "baseapp")
}

// Start starts the baseapp.
func (b *BaseApp) Start(ctx context.Context) error {
	b.Logger().Info("attempting to start")
	defer b.Logger().Info("successfully started")

	// Start the job manager and the producers.
	b.jobMgr.Start(ctx)
	b.jobMgr.RunProducers(ctx)

	if b.svr == nil {
		b.Logger().Info("no HTTP server registered, skipping")
	} else {
		go b.svr.Start(ctx)
	}

	return nil
}

// Stop stops the baseapp.
func (b *BaseApp) Stop() {
	b.Logger().Info("attempting to stop")
	defer b.Logger().Info("successfully stopped")

	b.jobMgr.Stop()
	if b.svr != nil {
		b.svr.Stop()
	}
}

// Chain returns the eth client for the baseapp.
func (b *BaseApp) Chain() eth.Client {
	return b.chain
}
