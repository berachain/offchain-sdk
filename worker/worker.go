package worker

import "github.com/alitto/pond"

type Pool struct {
	*pond.WorkerPool
}

func NewPool(name string, size int, capacity int) *Pool {
	return &Pool{
		WorkerPool: pond.New(size, capacity),
	}
}

// AddTask adds a task to the pool.
func (p *Pool) AddTask(exec Executor) {
	p.Submit(
		func() { exec.Execute() })
}
