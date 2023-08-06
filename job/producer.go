package job

import "context"

// Producer defines the interface of a function that can be spawned on a thread to
// produce jobs. This function should run until the provided context is.
type Producer[T Basic] func(ctx context.Context, pool WorkerPool, job T)
