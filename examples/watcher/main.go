package main

import (
	"fmt"
	"os"

	"github.com/berachain/offchain-sdk/baseapp"
	"github.com/berachain/offchain-sdk/cmd"
	"github.com/berachain/offchain-sdk/job"
)

func main() {
	appBuilder := baseapp.NewAppBuilder("watcher")

	appBuilder.RegisterJob(
		job.EthEventJob{},
	)

	if err := cmd.BuildBasicRootCmd(appBuilder).Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Whoops. There was an error while executing your CLI '%s'", err)
		os.Exit(1)
	}
}
