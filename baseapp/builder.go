package baseapp

import (
	"errors"

	"github.com/berachain/offchain-sdk/v2/client/eth"
	"github.com/berachain/offchain-sdk/v2/job"
	"github.com/berachain/offchain-sdk/v2/log"
	"github.com/berachain/offchain-sdk/v2/server"
	"github.com/berachain/offchain-sdk/v2/telemetry"

	ethdb "github.com/ethereum/go-ethereum/ethdb"
)

// AppBuilder is a builder for an app.
type AppBuilder struct {
	appName   string
	jobs      []job.Basic
	db        ethdb.KeyValueStore
	ethClient eth.Client
	svr       *server.Server
	metrics   telemetry.Metrics
	logger    log.Logger
}

// NewAppBuilder creates a new app builder.
func NewAppBuilder(appName string, logger log.Logger) *AppBuilder {
	return &AppBuilder{
		appName: appName,
		jobs:    []job.Basic{},
		metrics: telemetry.NewNoopMetrics(),
		logger:  logger,
	}
}

// AppName returns the name of the app.
func (ab *AppBuilder) AppName() string {
	return ab.appName
}

// AppName sets the name of the app.
func (ab *AppBuilder) RegisterJob(job job.Basic) {
	ab.jobs = append(ab.jobs, job)
}

// RegisterDB registers the db.
func (ab *AppBuilder) RegisterDB(db ethdb.KeyValueStore) {
	ab.db = db
}

// RegisterMetrics registers the metrics.
func (ab *AppBuilder) RegisterMetrics(cfg *telemetry.Config) error {
	var err error
	ab.metrics, err = telemetry.NewMetrics(cfg)
	if err != nil {
		return err
	}

	// Enable metrics on eth client only if it is an instance of ChainProviderImpl.
	chainProvider, ok := ab.ethClient.(*eth.ChainProviderImpl)
	if ok {
		chainProvider.EnableMetrics(ab.metrics)
	}
	return nil
}

// RegisterHTTPServer registers the http server.
func (ab *AppBuilder) RegisterHTTPServer(svr *server.Server) {
	ab.svr = svr
}

// RegisterHTTPHandler registers a HTTP handler.
func (ab *AppBuilder) RegisterHTTPHandler(handler *server.Handler) error {
	if ab.svr == nil {
		return errors.New("must enable the HTTP server to register a handler")
	}

	ab.svr.RegisterHandler(handler)
	return nil
}

// RegisterMiddleware registers a middleware to the HTTP server.
func (ab *AppBuilder) RegisterMiddleware(m server.Middleware) error {
	if ab.svr == nil {
		return errors.New("must enable the HTTP server to register a middleware")
	}

	ab.svr.RegisterMiddleware(m)
	return nil
}

// RegisterEthClient registers the eth client.
// TODO: update this to connection pool on baseapp and context gets one for running
func (ab *AppBuilder) RegisterEthClient(ethClient eth.Client) {
	ab.ethClient = ethClient
}

// BuildApp builds the app.
func (ab *AppBuilder) BuildApp() *BaseApp {
	return New(
		ab.appName,
		ab.logger,
		ab.ethClient,
		ab.jobs,
		ab.db,
		ab.svr,
		ab.metrics,
	)
}
