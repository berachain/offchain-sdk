package types

import (
	"context"

	"github.com/berachain/offchain-sdk/worker"
)

type Executable interface {
	Execute(context.Context, any) (any, error)
}

// Payload encapsulates a job and its input into a neat package to
// be executed by another thread.
type Payload struct {
	// Pass basic job
	job Executable

	// ctx is the context of the job.
	ctx context.Context

	// args is the input function arguments.
	args any
}

// NewPayload creates a new payload to send to a worker.
func NewPayload(ctx context.Context, job Executable, args any) *Payload {
	return &Payload{
		job:  job,
		ctx:  ctx,
		args: args,
	}
}

// Execute executes the job and returns the result.
// Todo: decouple from the worker package.
func (p Payload) Execute() worker.Resultor {
	res, err := p.job.Execute(p.ctx, p.args)
	return &Resultor{res: res, err: err}
}
