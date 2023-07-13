package worker

import (
	"errors"
	"fmt"
)

type Executor interface {
	Execute() any
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
}

// NewWorker creates a new worker
func NewWorker(
	newPayload chan Executor,
	newRes chan Resulter,
) *worker {
	return &worker{
		newPayload: newPayload,
		newRes:     newRes,
	}
}

// Start starts the worker
func (w *worker) Start() error {
	for {
		fmt.Println("WAITING FOR JOB")
		select {
		case executor, ok := <-w.newPayload:
			if !ok {
				return errors.New("bad payload")
			}
			// fan-in job execution multiplexing results into the results channel
			w.newRes <- executor.Execute()
		case <-w.stop:
			fmt.Printf("cancelled worker")
			return errors.New("cancelled worker")
		}
	}
}

// Stop stops the worker
func (w *worker) Stop() {
	w.stop <- struct{}{}
}
