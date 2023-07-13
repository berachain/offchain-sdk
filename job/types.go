package job

import (
	"context"

	"github.com/berachain/offchain-sdk/worker"
)

type HasContext[I any] interface {
	Ctx() context.Context
	Args() any
	Execute(context.Context, any) any
}

type Basic[I, O any] interface {
	Execute(context.Context, I) (O, error)
}

// type Conditional[A, O any] interface {
// 	Basic[A, O]
// 	Condition() bool
// }

// Executor encapsulates a job and its input into a neat package to
// be executed by another thread.
type Executor[I, O any] struct {
	// Pass basic job
	Job Basic[I, O]

	// And inputs
	ctx  context.Context
	args I
}

// Result encapsulates the result of a job execution.
type Resulter[O any] struct {
	res O
	err error
}

func (r Resulter[O]) Result() O {
	return r.res
}

func (r Resulter[O]) Error() error {
	return r.err
}

func (p *Executor[I, O]) Execute() worker.Resulter {
	res, err := p.Job.Execute(p.ctx, p.args)
	return Resulter[O]{res: res, err: err}
}
