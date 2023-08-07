package job

// TODO: decouple the job package from the worker package somehow.
import "github.com/berachain/offchain-sdk/worker"

type Payload interface {
	Execute() worker.Resultor
}

type WorkerPool interface {
	SubmitJob(worker.Payload) error
}

type WorkerShim struct {
	worker.Pool
}

func (ws *WorkerShim) SubmitJob(p Payload) error {
	ws.Pool.SubmitJob(p)
	return nil
}
