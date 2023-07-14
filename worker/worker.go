package worker

import (
	"sync"

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

	// wg is used to wait for the worker to stop.
	wg *sync.WaitGroup
}

// NewWorker creates a new worker.
func newWorker(
	newExecutor chan Executor,
	newRes chan Resultor,
	logger log.Logger,
	wg *sync.WaitGroup,
) *worker {
	return &worker{
		logger:      logger,
		newExecutor: newExecutor,
		newRes:      newRes,
		wg:          wg,
	}
}

// Start starts the worker.
func (w *worker) Start() {
	// Manage stopping the worker.
	w.stop = make(chan struct{}, 1)
	w.wg.Add(1)
	defer close(w.stop)

	w.logger.Info("starting worker")
	for {
		select {
		case <-w.stop:
			w.logger.Info("stopping worker")
			w.wg.Done()
			return
		case executor, ok := <-w.newExecutor:
			if !ok {
				w.logger.Error("worker stopped because of error")
				w.wg.Done()
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
