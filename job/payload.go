package job

import (
	"context"

	"github.com/berachain/offchain-sdk/worker"
)

// Payload encapsulates a job and its input into a neat package to
// be executed by another thread.
type Payload struct {
	// Pass basic job
	Job Basic

	// ctx is the context of the job.
	ctx context.Context

	// args is the input function arguments.
	args any
}

// NewPayload creates a new payload to send to a worker.
func NewPayload(ctx context.Context, job Basic, args any) *Payload {
	return &Payload{
		Job:  job,
		ctx:  ctx,
		args: args,
	}
}

// Execute executes the job and returns the result.
func (p Payload) Execute() worker.Resultor {
	res, err := p.Job.Execute(p.ctx, p.args)
	return &Resultor{res: res, err: err}
}
