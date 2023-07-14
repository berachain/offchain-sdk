package main

import (
	"fmt"
	"os"

	"github.com/berachain/offchain-sdk/baseapp"
	"github.com/berachain/offchain-sdk/cmd"
	sdk "github.com/berachain/offchain-sdk/types"
)

// Chain is building blocks.
type EthEventJob struct{}

func (j EthEventJob) Execute(ctx sdk.Context, args any) (any, error) {
	fmt.Println("HELLO BLOCK 10 OR HIGHER")
	return false, nil
}

func (j EthEventJob) Condition(ctx sdk.Context) bool {
	fmt.Println("CHECKING CONDITION")
	chain := ctx.Chain()
	block, err := chain.CurrentBlock()
	if err != nil || block.NumberU64() < 10 {
		return false
	}
	return true
}

func main() {
	appBuilder := baseapp.NewAppBuilder("watcher")

	appBuilder.RegisterJob(
		EthEventJob{},
	)

	if err := cmd.BuildBasicRootCmd(appBuilder).Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Whoops. There was an error while executing your CLI '%s'", err)
		os.Exit(1)
	}
}
