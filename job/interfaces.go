package job

import "github.com/berachain/offchain-sdk/worker"

type WorkerPool interface {
	SubmitJob(worker.Payload) error
}
