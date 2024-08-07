package config

import (
	"github.com/berachain/offchain-sdk/v2/client/eth"
	"github.com/berachain/offchain-sdk/v2/log"
	"github.com/berachain/offchain-sdk/v2/server"
)

// Reader defines an interface for reading in configuration data.
type Reader[C any] func(string, *C) error

// Config represents a configuration for the application + the offchain-sdk
// pieces.
type Config[C any] struct {
	// Application specific config
	App C

	// EthClient config
	ConnectionPool eth.ConnectionPoolConfig

	// Server Config
	Server server.Config

	// Log Config
	Log log.Config
}
