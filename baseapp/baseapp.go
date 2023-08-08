package baseapp

import (
	"context"

	"github.com/berachain/offchain-sdk/client/eth"
	"github.com/berachain/offchain-sdk/job"
	"github.com/berachain/offchain-sdk/log"
	"github.com/berachain/offchain-sdk/server"
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
	jobMgr *Manager

	// ethClient is the client for communicating with the chain
	ethClient eth.Client

	// db KV store
	db ethdb.KeyValueStore

	// svr is the server for the baseapp.
	svr *server.Server
}

// New creates a new baseapp.
func New(
	name string,
	logger log.Logger,
	ethClient eth.Client,
	jobs []job.Basic,
	db ethdb.KeyValueStore,
	svr *server.Server,
) *BaseApp {
	return &BaseApp{
		name:      name,
		logger:    logger,
		ethClient: ethClient,
		jobMgr: NewManager(
			jobs,
		),
		db:  db,
		svr: svr,
	}
}

// Logger returns the logger for the baseapp.
func (b *BaseApp) Logger() log.Logger {
	return b.logger.With("namespace", b.name+"-app")
}

// Start starts the baseapp.
func (b *BaseApp) Start(ctx context.Context) error {
	b.Logger().Info("starting app")

	// Wrap the context in sdk.Context in order to attach our clients, logger and db.
	// TODO: is this bad practice we are just stealing from the cosmos sdk?
	ctx = sdk.NewContext(
		ctx,
		b.ethClient,
		b.logger,
		b.db,
	)

	// Start the job manager and the producers.
	b.jobMgr.Start(ctx)
	b.jobMgr.RunProducers(ctx)

	// Start the server.
	go b.svr.Start()

	return nil
}

// Stop stops the baseapp.
func (b *BaseApp) Stop() {
	b.Logger().Info("stopping app")
	b.jobMgr.Stop()
	b.svr.Stop()
}
