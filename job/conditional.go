package job

import (
	"context"
	"time"
)

// Conditional represents a conditional job.
type Conditional interface {
	Polling
	Condition(ctx context.Context) bool
}

// ConditionalProducer produces a job when the condition is met.
func ConditionalProducer(ctx context.Context, pool WorkerPool, cj Conditional) error {
	for {
		select {
		// If the context is cancelled, return.
		case <-ctx.Done():
			return ctx.Err()
		default:
			// Sleep for a period of time.
			time.Sleep(cj.IntervalTime(ctx) * time.Millisecond)

			// Check if the condition is true.
			if cj.Condition(ctx) {
				// If true add a job
				pool.AddJob(NewPayload(ctx, cj, nil))
			}
		}
	}
}
