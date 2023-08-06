package main

import (
	"fmt"
	"os"

	"github.com/berachain/offchain-sdk/baseapp"
	"github.com/berachain/offchain-sdk/client/eth"
	"github.com/berachain/offchain-sdk/cmd"
	"github.com/berachain/offchain-sdk/examples/listener/config"
	ljobs "github.com/berachain/offchain-sdk/examples/listener/jobs"
	jobs "github.com/berachain/offchain-sdk/x/jobs"
	memdb "github.com/ethereum/go-ethereum/ethdb/memorydb"
)

const (
	configPath = "config/.env"
)

// This example shows how to watch for an event on the Ethereum blockchain.
// The event is defined in the smart contract at: 0x18Df82C7E422A42D47345Ed86B0E935E9718eBda
// The event is called: NumberChanged(uint256)
// The event is emitted when the number is changed in the smart contract.
// The event is watched by the offchain-sdk and when it is emitted, the execution function is called.
func main() {
	appBuilder := baseapp.NewAppBuilder("listener")
	appConfig := config.LoadConfig(configPath)

	// regsiter jobs
	appBuilder.RegisterJob(
		jobs.NewEthSub(
			&ljobs.Listener{},
			appConfig.AddressToListen,
			appConfig.EventName,
		),
	)

	appBuilder.RegisterJob(
		jobs.NewEthSub(
			&ljobs.DbWriter{},
			appConfig.AddressToListen,
			appConfig.EventName,
		),
	)

	appBuilder.RegisterJob(
		&ljobs.Poller{},
	)

	// register db
	appBuilder.RegisterDB(memdb.New())

	// register ethClient
	// TODO: move to connection pool
	ethConfig := eth.LoadConfig(configPath)
	ethClient := eth.NewClient(&ethConfig)
	appBuilder.RegisterEthClient(ethClient)

	// build command and run the app
	if err := cmd.BuildBasicRootCmd(appBuilder).Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Whoops. There was an error while executing your CLI '%s'", err)
		os.Exit(1)
	}
}
