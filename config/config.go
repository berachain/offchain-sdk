package config

import "github.com/berachain/offchain-sdk/client/eth"

// Reader defines an interface for reading in configuration data.
type Reader[C any] func(string, *C) error

// Config represents a configuration for the application + the offchain-sdk
// pieces.
type Config[C any] struct {
	// Application specific config
	App C

	// Offchain-SDK config
	Eth eth.Config
}
