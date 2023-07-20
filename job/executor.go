package job

import (
	"context"
)

// Executor encapsulates a job and its input into a neat package to
// be executed by another thread.
type Executor struct {
	// Pass basic job
	Job Basic

	// ctx is the context of the job.
	ctx context.Context

	// args is the input function arguments.
	args any
}

func NewExecutor(ctx context.Context, job Basic, args any) *Executor {
	return &Executor{
		Job:  job,
		ctx:  ctx,
		args: args,
	}
}

// Execute executes the job and returns the result.
func (p Executor) Execute() *Resultor {
	res, err := p.Job.Execute(p.ctx, p.args)
	return &Resultor{res: res, err: err}
}
