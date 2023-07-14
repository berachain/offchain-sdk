package worker

import (
	"fmt"

	"github.com/berachain/offchain-sdk/log"
)

// worker is a worker thread that executes jobs.
type worker struct {
	id uint32
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
	id uint32,
	newExecutor chan Executor,
	newRes chan Resultor,
	logger log.Logger,
) *worker {
	return &worker{
		id:          id,
		logger:      logger,
		newExecutor: newExecutor,
		newRes:      newRes,
	}
}

// Logger returns the logger for the worker.
func (w *worker) Logger() log.Logger {
	return w.logger.With("namespace", fmt.Sprintf("worker-%d", w.id))
}

// Start starts the worker.
func (w *worker) Start() {
	// Manage stopping the worker.
	w.stop = make(chan struct{}, 1)
	defer close(w.stop)

	w.Logger().Info("starting")
	for {
		select {
		case <-w.stop:
			w.Logger().Info("stopping")
			return
		case executor, ok := <-w.newExecutor:
			if !ok {
				w.Logger().Info("payload closed")
				return
			}
			w.Logger().Info("executing job")
			w.newRes <- executor.Execute()
		}
	}
}

// Stop stops the worker.
func (w *worker) Stop() {
	w.Logger().Info("triggering worker to stop")
	w.stop <- struct{}{}
}
