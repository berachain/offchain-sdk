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
	// name is the name of the application.
	name string

	// logger is the logger for the baseapp.
	logger log.Logger

	// jobMgr manages jobs within the baseapp.
	jobMgr JobManager

	// svr is the server for the baseapp.
	svr Server
}

// JobManager defines the interface for a job manager.
type JobManager interface {
	Start(ctx context.Context)
	RunProducers(ctx context.Context)
	Stop()
}

// Server defines the interface for a server.
type Server interface {
	Start(ctx context.Context) error
	Stop()
}

// New creates a new BaseApp instance.
func New(
	name string,
	logger log.Logger,
	ethClient eth.Client,
	jobs []job.Basic,
	db ethdb.KeyValueStore,
	svr Server,
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

// Logger returns a namespaced logger for the baseapp.
func (b *BaseApp) Logger() log.Logger {
	return b.logger.With("namespace", "baseapp")
}

// Start starts the baseapp and all its components.
func (b *BaseApp) Start(ctx context.Context) error {
	b.Logger().Info("Starting baseapp")
	defer b.Logger().Info("Baseapp started successfully")

	// Start the job manager and producers.
	b.jobMgr.Start(ctx)
	b.jobMgr.RunProducers(ctx)

	// Start the server if it's provided.
	if b.svr != nil {
		if err := b.startServer(ctx); err != nil {
			return err
		}
	} else {
		b.Logger().Info("No HTTP server registered, skipping")
	}

	return nil
}

// Stop gracefully stops the baseapp and its components.
func (b *BaseApp) Stop() {
	b.Logger().Info("Stopping baseapp")
	defer b.Logger().Info("Baseapp stopped successfully")

	b.jobMgr.Stop()
	if b.svr != nil {
		b.svr.Stop()
	}
}

// startServer starts the HTTP server in a separate goroutine.
func (b *BaseApp) startServer(ctx context.Context) error {
	b.Logger().Info("Starting HTTP server")
	errChan := make(chan error)

	go func() {
		if err := b.svr.Start(ctx); err != nil {
			errChan <- err
		}
		close(errChan)
	}()

	select {
	case err := <-errChan:
		if err != nil {
			b.Logger().Error("Failed to start HTTP server", "error", err)
			return err
		}
	case <-ctx.Done():
		b.Logger().Warn("Server start canceled", "reason", ctx.Err())
		return ctx.Err()
	}

	return nil
}
