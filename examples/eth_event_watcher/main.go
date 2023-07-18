package main

import (
	"context"
	"fmt"
	"os"

	"github.com/berachain/offchain-sdk/baseapp"
	"github.com/berachain/offchain-sdk/cmd"
	"github.com/berachain/offchain-sdk/job"
	"github.com/berachain/offchain-sdk/log"
	"github.com/ethereum/go-ethereum/common"
)

func main() {
	appBuilder := baseapp.NewAppBuilder("watcher", "")
	exec := func(ctx context.Context, args any) (any, error) {
		logger := log.NewLogger(os.Stdout, "execution-logger")
		logger.Info("executing event function", "args", args)
		return nil, nil
	}
	appBuilder.RegisterJob(
		job.NewEthSub(common.HexToAddress("0x18Df82C7E422A42D47345Ed86B0E935E9718eBda"), "NumberChanged(uint256)", exec),
	)

	if err := cmd.BuildBasicRootCmd(appBuilder).Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Whoops. There was an error while executing your CLI '%s'", err)
		os.Exit(1)
	}
}
