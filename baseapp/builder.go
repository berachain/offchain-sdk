package baseapp

import (
	"errors"

	"github.com/berachain/offchain-sdk/client/eth"
	"github.com/berachain/offchain-sdk/job"
	"github.com/berachain/offchain-sdk/log"
	"github.com/berachain/offchain-sdk/server"
	"github.com/berachain/offchain-sdk/telemetry"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	ethdb "github.com/ethereum/go-ethereum/ethdb"
)

// AppBuilder is used to construct an application with the necessary components.
type AppBuilder struct {
	appName   string
	jobs      []job.Basic
	db        ethdb.KeyValueStore
	ethClient eth.Client
	svr       *server.Server
	metrics   telemetry.Metrics
}

// NewAppBuilder creates a new instance of AppBuilder with a given app name.
func NewAppBuilder(appName string) *AppBuilder {
	return &AppBuilder{
		appName: appName,
		jobs:    []job.Basic{},
		metrics: telemetry.NewNoopMetrics(), // Default to no-op metrics.
	}
}

// AppName returns the name of the app.
func (ab *AppBuilder) AppName() string {
	return ab.appName
}

// RegisterJob registers a new job in the app.
func (ab *AppBuilder) RegisterJob(job job.Basic) {
	ab.jobs = append(ab.jobs, job)
}

// RegisterDB registers the application's database connection.
func (ab *AppBuilder) RegisterDB(db ethdb.KeyValueStore) {
	ab.db = db
}

// RegisterMetrics configures and registers the app's metrics system.
func (ab *AppBuilder) RegisterMetrics(cfg *telemetry.Config) error {
	var err error
	ab.metrics, err = telemetry.NewMetrics(cfg)
	if err != nil {
		return err
	}

	// Enable metrics on eth client only if it is an instance of ChainProviderImpl.
	if chainProvider, ok := ab.ethClient.(*eth.ChainProviderImpl); ok {
		chainProvider.EnableMetrics(ab.metrics)
	}
	return nil
}

// RegisterHTTPServer enables and configures the HTTP server for the app.
func (ab *AppBuilder) RegisterHTTPServer(svr *server.Server) {
	ab.svr = svr
}

// RegisterHTTPHandler registers an HTTP handler to the app's server.
func (ab *AppBuilder) RegisterHTTPHandler(handler *server.Handler) error {
	if ab.svr == nil {
		return errors.New("HTTP server must be enabled before registering a handler")
	}

	ab.svr.RegisterHandler(handler)
	return nil
}

// RegisterMiddleware registers a middleware function to the app's HTTP server.
func (ab *AppBuilder) RegisterMiddleware(m server.Middleware) error {
	if ab.svr == nil {
		return errors.New("HTTP server must be enabled before registering middleware")
	}

	ab.svr.RegisterMiddleware(m)
	return nil
}

// RegisterPrometheusTelemetry registers a Prometheus-compatible HTTP handler for metrics.
func (ab *AppBuilder) RegisterPrometheusTelemetry() error {
	if ab.svr == nil {
		return errors.New("HTTP server must be enabled before registering Prometheus telemetry")
	}

	// Register the Prometheus metrics handler at the "/metrics" endpoint.
	return ab.RegisterHTTPHandler(&server.Handler{
		Path:    "/metrics",
		Handler: promhttp.Handler(),
	})
}

// RegisterEthClient registers the Ethereum client to be used by the app.
// TODO: Update this to utilize a connection pool for the Ethereum client, and ensure context management 
// allows for requesting a specific connection for each operation.
func (ab *AppBuilder) RegisterEthClient(ethClient eth.Client) {
	ab.ethClient = ethClient
}

// BuildApp finalizes the app setup and returns a new instance of the app.
func (ab *AppBuilder) BuildApp(logger log.Logger) *BaseApp {
	return New(
		ab.appName,
		logger,
		ab.ethClient,
		ab.jobs,
		ab.db,
		ab.svr,
		ab.metrics,
	)
}
