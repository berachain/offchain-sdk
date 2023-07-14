package main

import (
	"context"
	"os"
	"os/signal"

	"github.com/berachain/offchain-sdk/job"
	"github.com/berachain/offchain-sdk/log"
	"github.com/berachain/offchain-sdk/worker"
)

// MyMockJob is a mock job.
type MyMockJob struct{}

// Execute executes the job and returns the result.
func (m MyMockJob) Execute(_ context.Context, i int64) (int64, error) {
	return i + 69, nil //nolint:gomnd // Mock job.
}

const numWorkers = 10

func main() {
	// Handle os.Signal for graceful shutdown
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)
	logger := log.NewLogger(os.Stdout, "thread")
	workerPool := worker.NewPool(numWorkers, logger)

	// Start the pool
	workerPool.Start()

	// Add 1000 tasks to the pool
	for i := 0; i < 1000; i++ {
		workerPool.AddTask(job.Executor[int64, int64]{
			Job: MyMockJob{}})
	}

	// Wait for a signal to stop
	for {
		select {
		case res := <-workerPool.RespChan():
			log.NewLogger(os.Stdout, "results").Info("result", "result", res.Result())
		case <-signalChan:
			workerPool.Stop()
			return
		}
	}
}
