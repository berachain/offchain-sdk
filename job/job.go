package job

import (
	sdk "github.com/berachain/offchain-sdk/types"
)

// Basic represents a basic job.
type Basic interface {
	Execute(sdk.Context, any) (any, error)
}

// Conditional represents a conditional job.
type Conditional interface {
	Basic
	Condition(ctx sdk.Context) bool
}
