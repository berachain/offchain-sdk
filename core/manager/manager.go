package manager

import (
	"time"
)

type BasicJob interface {
	Execute() error
}

type ConditionalJob interface {
	BasicJob
	Condition() bool
}

// Subber needs a base job
type subber struct {
	job  BasicJob
	sub  chan struct{}
	stop chan struct{}
}

func NewSubber(job BasicJob) *subber {
	return &subber{
		job: job,
	}
}

// Start starts the subber
func (s *subber) Start() {
	for {
		select {
		case <-s.stop:
			return
		case <-s.sub:
			if err := s.job.Execute(); err != nil {
				return
			}
		}
	}
}

// Stop stops the subber
func (s *subber) Stop() {
	s.stop <- struct{}{}
	close(s.stop)
}

// Poller needs a conditional job

// Poller is a poller
type poller struct {
	job      ConditionalJob
	stop     chan struct{}
	interval time.Duration
}

// NewPoller creates a new poller
func NewPoller(job ConditionalJob, interval time.Duration) *poller {
	return &poller{
		job:  job,
		stop: make(chan struct{}),
	}
}

// Start starts the poller
func (p *poller) Start() {
	for {
		select {
		case <-p.stop:
			return
		default:
			if p.job.Condition() {
				p.job.Execute()
			}
		}

		// Sleep for the interval
		time.Sleep(p.interval)
	}
}

// Stop stops the poller
func (p *poller) Stop() {
	p.stop <- struct{}{}
	close(p.stop)
}

type manager struct {
}

func NewManager() *manager {
	return &manager{}
}

func (m *manager) RegisterJob(job BasicJob) {

}
