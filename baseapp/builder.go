package baseapp

import (
	"github.com/berachain/offchain-sdk/client/eth"
	"github.com/berachain/offchain-sdk/log"
)

// AppBuilder is a builder for an app.
type AppBuilder struct {
	appName string
}

// NewAppBuilder creates a new app builder.
func NewAppBuilder(appName string) *AppBuilder {
	return &AppBuilder{
		appName: appName,
	}
}

// AppName returns the name of the app.
func (ab *AppBuilder) AppName() string {
	return ab.appName
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
	)
}
