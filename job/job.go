package job

import (
	"fmt"

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

// Chain is building blocks.
type EthEventJob struct{}

func (j EthEventJob) Execute(ctx sdk.Context, args any) (any, error) {
	fmt.Println("HELLO BLOCK 20 OR HIGHER")
	return false, nil
}

func (j EthEventJob) Condition(ctx sdk.Context) bool {
	fmt.Println("CHECKING CONDITION")
	chain := ctx.Chain()
	block, err := chain.CurrentBlock()
	if err != nil || block.NumberU64() < 20 {
		return false
	}
	return true
}
