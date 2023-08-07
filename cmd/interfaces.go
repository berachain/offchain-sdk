package cmd

import (
	"github.com/berachain/offchain-sdk/baseapp"
	"github.com/berachain/offchain-sdk/log"
)

// AppBuilder is a builder for an app. It follows a basic factory pattern.
type AppBuilder interface {
	AppName() string
	BuildApp(log.Logger) *baseapp.BaseApp
}
