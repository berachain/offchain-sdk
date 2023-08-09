package app

import (
	"github.com/berachain/offchain-sdk/cmd"
	"github.com/berachain/offchain-sdk/log"

	"github.com/berachain/offchain-sdk/baseapp"
	"github.com/berachain/offchain-sdk/examples/listener/config"
	ljobs "github.com/berachain/offchain-sdk/examples/listener/jobs"
	jobs "github.com/berachain/offchain-sdk/x/jobs"
	memdb "github.com/ethereum/go-ethereum/ethdb/memorydb"
)

// TODO: move cmd.App out of the cmd package.
// We must conform to the `App` interface.
var _ cmd.App[config.Config] = &ListenerApp{}

// ListenerApp shows how to watch for an event on the Ethereum blockchain.
// The event is defined in the smart contract at: 0x18Df82C7E422A42D47345Ed86B0E935E9718eBda
// The event is called: NumbreChanged(uint256)
// The event is emitted when the number is changed in the smart contract.
// The event is watched by the offchain-sdk and when it is emitted, the execution function is called.
type ListenerApp struct {
	*baseapp.BaseApp
}

// Name implements the `App` interface.
func (ListenerApp) Name() string {
	return "listener-app"
}

// Setup implements the `App` interface.
func (app *ListenerApp) Setup(
	ab cmd.AppBuilder,
	config config.Config,
	logger log.Logger,
) {
	// This job is subscribed to the `NumberChanged(uint256)` event.
	ab.RegisterJob(
		jobs.NewEthSub(
			&ljobs.Listener{}, // We embed a Basic job inside.
			config.Jobs.Sub.AddressToListen,
			config.Jobs.Sub.EventName,
		),
	)

	// This job is also subscribed to the `NumberChanged(uint256)` event.
	ab.RegisterJob(
		jobs.NewEthSub(
			&ljobs.DbWriter{},
			config.Jobs.Sub.AddressToListen,
			config.Jobs.Sub.EventName,
		),
	)

	// This job is querying the chain on a 1 second time interval.
	ab.RegisterJob(
		&ljobs.Poller{
			Interval: config.Jobs.Poller.Interval,
		},
	)

	// We register a database with our app.
	ab.RegisterDB(memdb.New())

	// And then we setup everything by calling `BuildApp`.
	app.BaseApp = ab.BuildApp(logger)
}
