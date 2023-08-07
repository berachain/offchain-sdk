package main

import (
	"fmt"
	"os"

	"github.com/berachain/offchain-sdk/baseapp"
	"github.com/berachain/offchain-sdk/cmd"
	ljobs "github.com/berachain/offchain-sdk/examples/listener/jobs"
	jobs "github.com/berachain/offchain-sdk/x/jobs"
	memdb "github.com/ethereum/go-ethereum/ethdb/memorydb"
)

// This example shows how to watch for an event on the Ethereum blockchain.
// The event is defined in the smart contract at: 0x18Df82C7E422A42D47345Ed86B0E935E9718eBda
// The event is called: NumberChanged(uint256)
// The event is emitted when the number is changed in the smart contract.
// The event is watched by the offchain-sdk and when it is emitted, the execution function is called.
func main() {
	appBuilder := baseapp.NewAppBuilder("listener")

	// TODO, proper config thing.
	eventName := "NumberChanged(uint256)"
	addrToListen := "0x5793a71D3eF074f71dCC21216Dbfd5C0e780132c"

	// regsiter jobs
	appBuilder.RegisterJob(
		jobs.NewEthSub(
			&ljobs.Listener{},
			addrToListen,
			eventName,
		),
	)

	appBuilder.RegisterJob(
		jobs.NewEthSub(
			&ljobs.DbWriter{},
			addrToListen,
			eventName,
		),
	)

	appBuilder.RegisterJob(
		&ljobs.Poller{},
	)

	// register db
	appBuilder.RegisterDB(memdb.New())

	// build command and run the app
	if err := cmd.BuildBasicRootCmd(appBuilder).Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Whoops. There was an error while executing your CLI '%s'", err)
		os.Exit(1)
	}
}
