package baseapp

import (
	"os"

	"github.com/berachain/offchain-sdk/client/eth"
	"github.com/berachain/offchain-sdk/log"
	"github.com/berachain/offchain-sdk/worker"
)

// BaseApp is the base application.
type BaseApp struct {
	// name is the name of the application
	name string

	// logger is the logger for the baseapp.
	logger log.Logger

	// contains filtered or unexported fields
	ethCfg eth.Config

	// worker pool
	workerPool worker.Pool
}

// New creates a new baseapp.
func New(
	name string,
	logger log.Logger,
	ethCfg *eth.Config) *BaseApp {
	return &BaseApp{
		name:   name,
		logger: log.NewBlankLogger(os.Stdout),
		ethCfg: *ethCfg,
		workerPool: worker.NewPool(
			"main",
			16, //nolint:gomnd // hardcode 16 workers for now
			logger,
		),
	}
}

// Logger returns the logger for the baseapp.
func (b *BaseApp) Logger() log.Logger {
	return b.logger.With("namespace", "baseapp")
}

// Start starts the baseapp.
func (b *BaseApp) Start() {
	b.Logger().Info("starting app")
	b.workerPool.Start()
}

// Stop stops the baseapp.
func (b *BaseApp) Stop() {
	b.Logger().Info("stopping app")
	b.workerPool.Stop()
}
