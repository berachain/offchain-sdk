package baseapp

import (
	"context"

	"cosmossdk.io/log"
	"github.com/berachain/offchain-sdk/client/eth"
	"github.com/berachain/offchain-sdk/job"
	sdk "github.com/berachain/offchain-sdk/types"
)

// BaseApp is the base application.
type BaseApp struct {
	// name is the name of the application
	name string

	// logger is the logger for the baseapp.
	logger log.Logger

	// contains filtered or unexported fields
	ethCfg eth.Config

	// jobMgr
	jobMgr *JobManager

	// ethClient is the client for communicating with the chain
	ethClient eth.Client
}

// New creates a new baseapp.
func New(
	name string,
	logger log.Logger,
	ethCfg *eth.Config,
	jobs []job.Basic,
) *BaseApp {
	return &BaseApp{
		name:   name,
		logger: logger,
		ethCfg: *ethCfg,
		ethClient: eth.NewClient(
			logger,
			ethCfg,
		),
		jobMgr: NewJobManager(
			name,
			logger,
			jobs,
		),
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
			eth.NewClient(b.logger, &b.ethCfg),
		),
		b.Logger(),
	)
	b.jobMgr.executionPool.Start()
	b.jobMgr.Start(*ctx)
}

// Stop stops the baseapp.
func (b *BaseApp) Stop() {
	b.Logger().Info("stopping app")
	b.jobMgr.executionPool.Stop()
}

// 1 ClientRouter -> 1 ConnectionPool -> N EthClients

// ClientRouter is the same interface as the EthClient
// It basically just grabs a client makes the call, returns the result and then puts the ethclient back into the pool
