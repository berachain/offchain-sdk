package baseapp

import (
	"context"
	"os"

	"github.com/berachain/offchain-sdk/client/eth"
	"github.com/berachain/offchain-sdk/job"
	"github.com/berachain/offchain-sdk/log"
	sdk "github.com/berachain/offchain-sdk/types"
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

	// ethClient is the client for communicating with the chain
	ethClient eth.Client

	// db KV store
	db ethdb.KeyValueStore
}

// New creates a new baseapp.
func New(
	name string,
	logger log.Logger,
	ethClient eth.Client,
	jobs []job.Basic,
	db ethdb.KeyValueStore,
) *BaseApp {
	return &BaseApp{
		name:      name,
		logger:    log.NewBlankLogger(os.Stdout),
		ethClient: ethClient,
		jobMgr: NewJobManager(
			name,
			logger,
			jobs,
		),
		db: db,
	}
}

// Logger returns the logger for the baseapp.
func (b *BaseApp) Logger() log.Logger {
	return b.logger.With("namespace", b.name+"-app")
}

// Start starts the baseapp.
func (b *BaseApp) Start() {
	b.Logger().Info("starting app")

	// TODO: create a new context for every job request / creation.
	ctx := sdk.NewContext(
		context.Background(),
		eth.NewContextualClient(
			context.Background(),
			b.ethClient,
		),
		b.Logger(),
		b.db,
	)
	b.jobMgr.executionPool.Start()
	b.jobMgr.Start(*ctx)
}

// Stop stops the baseapp.
func (b *BaseApp) Stop() {
	b.Logger().Info("stopping app")
	b.jobMgr.executionPool.Stop()
}
