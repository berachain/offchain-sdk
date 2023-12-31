package baseapp

import (
	"github.com/berachain/offchain-sdk/client/eth"
	"github.com/berachain/offchain-sdk/job"
	"github.com/berachain/offchain-sdk/log"
	"github.com/berachain/offchain-sdk/server"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	ethdb "github.com/ethereum/go-ethereum/ethdb"
)

// AppBuilder is a builder for an app.
type AppBuilder struct {
	appName   string
	jobs      []job.Basic
	db        ethdb.KeyValueStore
	ethClient eth.Client
	svr       *server.Server
}

// NewAppBuilder creates a new app builder.
func NewAppBuilder(appName string) *AppBuilder {
	return &AppBuilder{
		appName: appName,
		jobs:    []job.Basic{},
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

// RegisterHTTPServer registers the http server.
func (ab *AppBuilder) RegisterHTTPServer(svr *server.Server) {
	ab.svr = svr

	// TODO: probably move
	ab.svr.RegisterHandler(
		server.Handler{Path: "/metrics", Handler: promhttp.Handler()},
	)
}

// RegisterDB registers the db.
func (ab *AppBuilder) RegisterHTTPHandler(handler server.Handler) {
	ab.svr.RegisterHandler(handler)
}

// RegisterEthClient registers the eth client.
// TODO: update this to connection pool on baseapp and context gets one for running
func (ab *AppBuilder) RegisterEthClient(ethClient eth.Client) {
	ab.ethClient = ethClient
}

// BuildApp builds the app.
func (ab *AppBuilder) BuildApp(
	logger log.Logger,
) *BaseApp {
	return New(
		ab.appName,
		logger,
		ab.ethClient,
		ab.jobs,
		ab.db,
		ab.svr,
	)
}
