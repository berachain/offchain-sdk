package job

import (
	"context"
)

// Basic represents a basic job.
type Basic[I, O any] interface {
	Execute(context.Context, I) (O, error)
}
