package eth

import "time"

const (
	defaultRPCTimeout          = 5 * time.Second
	defaultHealthCheckInterval = 5 * time.Second
)

type ConnectionPoolConfig struct {
	EthHTTPURLs         []string
	EthWSURLs           []string
	DefaultTimeout      time.Duration
	HealthCheckInterval time.Duration
}

func DefaultConnectPoolConfig() *ConnectionPoolConfig {
	return &ConnectionPoolConfig{
		EthHTTPURLs:         []string{"http://localhost:8545"},
		EthWSURLs:           []string{"ws://localhost:8546"},
		DefaultTimeout:      defaultRPCTimeout,
		HealthCheckInterval: defaultHealthCheckInterval,
	}
}
