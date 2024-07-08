package app

import (
	"time"

	"github.com/berachain/offchain-sdk/baseapp"
	coreapp "github.com/berachain/offchain-sdk/core/app"
	"github.com/berachain/offchain-sdk/examples/simple-metrics-app/config"
	"github.com/berachain/offchain-sdk/examples/simple-metrics-app/jobs"
	"github.com/berachain/offchain-sdk/log"
)

// We must conform to the `App` interface.
var _ coreapp.App[config.AppConfig] = &SimpleMetricsApp{}

// SimpleMetricsApp shows how to set up metrics on rpc methods.
type SimpleMetricsApp struct {
	*baseapp.BaseApp
}

// Name implements the `App` interface.
func (SimpleMetricsApp) Name() string {
	return "simple-metrics-app"
}

// Setup implements the `App` interface.
func (app *SimpleMetricsApp) Setup(
	ab coreapp.Builder,
	config config.AppConfig,
	logger log.Logger,
) {
	var err error

	// Set up metrics instance
	err = ab.RegisterMetrics(&config.Metrics)
	if err != nil {
		logger.Error("error setting up metrics", "error", err)
		return
	}

	// Spin up Prometheus HTTP server
	if config.Metrics.Prometheus.Enabled {
		if err = ab.RegisterPrometheusTelemetry(); err != nil {
			panic(err)
		}
	}

	// This job is querying the chain on a 1 second time interval.
	ab.RegisterJob(
		&jobs.Poller{
			Interval: time.Second,
		},
	)

	ab.RegisterJob(
		&jobs.Parser{
			Interval: time.Second * 5,
		},
	)

	// And then we setup everything by calling `BuildApp`.
	app.BaseApp = ab.BuildApp(logger)
}
