package worker

import (
	"github.com/berachain/offchain-sdk/log"
)

// worker is a worker thread that executes jobs.
type worker struct {
	// Gets jobs fed to it.
	newExecutor chan (Executor)
	// Feeds results onto a channel.
	newRes chan (Resultor)
	// Notify the worker to stop.
	stop chan struct{}
	// logger represents our logger
	logger log.Logger
}

// NewWorker creates a new worker.
func newWorker(
	newExecutor chan Executor,
	newRes chan Resultor,
	logger log.Logger,
) *worker {
	return &worker{
		logger:      logger,
		newExecutor: newExecutor,
		newRes:      newRes,
	}
}

// Start starts the worker.
func (w *worker) Start() {
	// Manage stopping the worker.
	w.stop = make(chan struct{}, 1)
	defer close(w.stop)

	w.logger.Info("starting worker")
	for {
		select {
		case <-w.stop:
			w.logger.Info("stopping worker")
			return
		case executor, ok := <-w.newExecutor:
			if !ok {
				w.logger.Info("payload closed")
				return
			}
			w.logger.Info("executing job")
			w.newRes <- executor.Execute()
		}
	}
}

// Stop stops the worker.
func (w *worker) Stop() {
	w.logger.Info("triggering worker to stop")
	w.stop <- struct{}{}
}
