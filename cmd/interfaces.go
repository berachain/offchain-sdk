package cmd

import (
	"github.com/berachain/offchain-sdk/baseapp"
	"github.com/berachain/offchain-sdk/client/eth"
	"github.com/berachain/offchain-sdk/job"
	"github.com/berachain/offchain-sdk/log"
	"github.com/ethereum/go-ethereum/ethdb"
)

// AppBuilder is a builder for an app. It follows a basic factory pattern.
type AppBuilder interface {
	AppName() string
	BuildApp(log.Logger) *baseapp.BaseApp
	RegisterEthClient(eth.Client)
	RegisterJob(job.Basic)
	RegisterDB(db ethdb.KeyValueStore)
}
