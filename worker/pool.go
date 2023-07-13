package worker

import (
	"fmt"
	"os"

	"github.com/berachain/offchain-sdk/log"
)

// pool is a pool of workers.
type pool struct {
	logger  log.Logger
	execCh  chan Executor
	resCh   chan Resulter
	workers []*worker
}

// NewPool creates a new pool of workers.
func NewPool(
	numWorkers uint64,
	logger log.Logger,
) *pool {
	// Intialize the pool.
	p := &pool{
		workers: make([]*worker, 0),
		execCh:  make(chan Executor),
		resCh:   make(chan Resulter),
	}

	// Iterate through the number of workers and create them.
	for i := uint64(0); i < numWorkers; i++ {
		w := NewWorker(
			p.execCh,
			p.resCh,
			log.NewLogger(os.Stdout, fmt.Sprintf("worker-%d", i)),
		)
		p.workers = append(p.workers, w)
	}

	return p
}

func (p *pool) Start() {
	// Start all the workers.
	for _, w := range p.workers {
		go w.Start()
	}
}

func (p *pool) Stop() {
	// Stop all the workers.
	for _, w := range p.workers {
		w.Stop()
		
	}
}

func (p *pool) AddTask(exec Executor) {
	p.execCh <- exec
}

func (p *pool) RespChan() chan Resulter {
	return p.resCh
}

func (p *pool) GetResult() Resulter {
	return <-p.resCh
}
