package worker

import (
	"errors"

	"github.com/berachain/offchain-sdk/log"
)

type Executor interface {
	Execute() Resulter
}

type Resulter interface {
	Result() any
	Error() error
}

// Subber needs a base job
type worker struct {
	// Gets jobs fed to it.
	newPayload chan (Executor)
	// Feeds results onto a channel.
	newRes chan (Resulter)
	// Notify the worker to stop.
	stop chan struct{}
	// logger represents our logger
	logger log.Logger
}

// NewWorker creates a new worker
func NewWorker(
	newPayload chan Executor,
	newRes chan Resulter,
	logger log.Logger,
) *worker {
	return &worker{
		logger:     logger,
		newPayload: newPayload,
		newRes:     newRes,
		stop:       make(chan struct{}),
	}
}

// Start starts the worker
func (w *worker) Start() error {
	w.logger.Info("starting worker")
	for {
		select {
		case executor, ok := <-w.newPayload:
			if !ok {
				return errors.New("bad payload")
			}
			// fan-in job execution multiplexing results into the results channel
			w.logger.Info("executing job")
			w.newRes <- executor.Execute()
		case <-w.stop:
			w.logger.Info("stopping worker")
			return errors.New("cancelled worker")
		}
	}
}

// Stop stops the worker
func (w *worker) Stop() {
	w.stop <- struct{}{}
}
