package baseapp

import (
	"github.com/berachain/offchain-sdk/client/eth"
	"github.com/berachain/offchain-sdk/job"
	"github.com/berachain/offchain-sdk/log"
)

// AppBuilder is a builder for an app.
type AppBuilder struct {
	appName    string
	jobs       []job.Basic
	configPath string
}

// NewAppBuilder creates a new app builder.
func NewAppBuilder(appName string, configPath string) *AppBuilder {
	return &AppBuilder{
		appName:    appName,
		jobs:       []job.Basic{},
		configPath: configPath,
	}
}

// AppName returns the name of the app.
func (ab *AppBuilder) AppName() string {
	return ab.appName
}

// ConfigPath returns the path to the config file.
func (ab *AppBuilder) ConfigPath() string {
	return ab.configPath
}

// AppName sets the name of the app.
func (ab *AppBuilder) RegisterJob(job job.Basic) {
	ab.jobs = append(ab.jobs, job)
}

// BuildApp builds the app.
func (ab *AppBuilder) BuildApp(
	logger log.Logger,
	ethCfg *eth.Config,
) *BaseApp {
	return New(
		ab.appName,
		logger,
		ethCfg,
		ab.jobs,
	)
}
