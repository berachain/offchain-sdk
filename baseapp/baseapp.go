package baseapp

import (
	"context"

	"github.com/berachain/offchain-sdk/client/eth"
	"github.com/berachain/offchain-sdk/job"
	"github.com/berachain/offchain-sdk/log"
	"github.com/berachain/offchain-sdk/server"
	"github.com/berachain/offchain-sdk/telemetry"
	ethdb "github.com/ethereum/go-ethereum/ethdb"
)

// BaseApp is the base application.
type BaseApp struct {
	name    string        // name of the application
	logger  log.Logger    // logger for the baseapp
	jobMgr  *JobManager   // job manager
	svr     *server.Server // server for the baseapp
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

	if b.svr != nil {
		go b.svr.Start(ctx)
	} else {
		b.Logger().Info("no HTTP server registered, skipping")
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
