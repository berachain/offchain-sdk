package main

import (
	"fmt"
	"os"

	"github.com/berachain/offchain-sdk/baseapp"
	"github.com/berachain/offchain-sdk/cmd"
	wjob "github.com/berachain/offchain-sdk/examples/eth_event_watcher/jobs"
	jobs "github.com/berachain/offchain-sdk/x/jobs"
)

// This example shows how to watch for an event on the Ethereum blockchain.
// The event is defined in the smart contract at: 0x18Df82C7E422A42D47345Ed86B0E935E9718eBda
// The event is called: NumberChanged(uint256)
// The event is emitted when the number is changed in the smart contract.
// The event is watched by the offchain-sdk and when it is emitted, the execution function is called.
func main() {
	appBuilder := baseapp.NewAppBuilder("watcher", "")

	appBuilder.RegisterJob(
		jobs.NewEthSub(
			&wjob.Watcher{},
			"0x18Df82C7E422A42D47345Ed86B0E935E9718eBda",
			"NumberChanged(uint256)",
		),
	)

	if err := cmd.BuildBasicRootCmd(appBuilder).Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Whoops. There was an error while executing your CLI '%s'", err)
		os.Exit(1)
	}
}
