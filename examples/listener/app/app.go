package app

import (
	"github.com/berachain/offchain-sdk/v2/baseapp"
	coreapp "github.com/berachain/offchain-sdk/v2/core/app"
	"github.com/berachain/offchain-sdk/v2/examples/listener/config"
	ljobs "github.com/berachain/offchain-sdk/v2/examples/listener/jobs"
	"github.com/berachain/offchain-sdk/v2/log"
	"github.com/berachain/offchain-sdk/v2/telemetry"
	"github.com/berachain/offchain-sdk/v2/tools/limiter"
	jobs "github.com/berachain/offchain-sdk/v2/x/jobs"

	memdb "github.com/ethereum/go-ethereum/ethdb/memorydb"
)

// TODO: move cmd.App out of the cmd package.
// We must conform to the `App` interface.
var _ coreapp.App[config.Config] = &ListenerApp{}

// ListenerApp shows how to watch for an event on the Ethereum blockchain.
// The event is defined in the smart contract at: 0x18Df82C7E422A42D47345Ed86B0E935E9718eBda
// The event is called: NumbreChanged(uint256)
// The event is emitted when the number is changed in the smart contract.
// The event is watched by the offchain-sdk and when emitted, the execution function is called.
type ListenerApp struct {
	*baseapp.BaseApp
	metrics telemetry.Metrics
}

// Name implements the `App` interface.
func (ListenerApp) Name() string {
	return "listener"
}

// Setup implements the `App` interface.
func (app *ListenerApp) Setup(
	ab coreapp.Builder,
	config config.Config,
	logger log.Logger,
) error {
	var err error

	// Set up metrics instance
	app.metrics, err = telemetry.NewMetrics(&config.Metrics)
	if err != nil {
		logger.Error("error setting up metrics", "error", err)
		return err
	}

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

	// This job is listening to the blocks on the chain.
	ab.RegisterJob(
		jobs.NewBlockHeaderWatcher(&ljobs.BlockWatcher{}),
	)

	// We register a database with our app.
	ab.RegisterDB(memdb.New())

	if config.RateLimit.Enabled {
		// We register a rate limiter with our app.
		rateLimiter := limiter.New(config.RateLimit)
		if err = ab.RegisterMiddleware(limiter.Middleware(rateLimiter)); err != nil {
			return err
		}
	}

	// And then we setup everything by calling `BuildApp`.
	app.BaseApp = ab.BuildApp()
	return nil
}
