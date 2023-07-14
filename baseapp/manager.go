package baseapp

import (
	"fmt"
	"os"
	"time"

	sdk "github.com/berachain/offchain-sdk/types"

	"github.com/berachain/offchain-sdk/job"
	"github.com/berachain/offchain-sdk/log"
	"github.com/berachain/offchain-sdk/worker"
)

type JobManager struct {
	// logger is the logger for the baseapp
	logger log.Logger

	// listening for conditions
	// conditionPool worker.Pool

	// list of jobs
	jobs []job.Conditional

	// worker pool
	executionPool worker.Pool
}

// New creates a new baseapp.
func NewJobManager(
	name string,
	logger log.Logger,
	jobs []job.Conditional,
) *JobManager {
	return &JobManager{
		logger: log.NewBlankLogger(os.Stdout),
		jobs:   jobs,
		executionPool: worker.NewPool(
			name+"-execution",
			16, //nolint:gomnd // hardcode 16 workers for now
			logger,
		),
	}
}

// Start
func (jm *JobManager) Start(ctx sdk.Context) {
	for _, j := range jm.jobs {
		fmt.Println("REEE")
		if basic, ok := j.(job.Conditional); ok {
			fmt.Println("REEE")
			go func() {
				for {
					time.Sleep(50 * time.Millisecond)
					if basic.Condition(ctx) {
						jm.executionPool.AddTask(job.NewExecutor(ctx, j, nil))
						return
					}
				}
			}()
		}
	}
}
