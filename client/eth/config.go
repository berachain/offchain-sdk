package eth

// Config is the configuration for the eth client.
type Config struct {
	EthHTTPURLs string
	EthWSURLs   string
}

func DefaultConfig() *Config {
	return &Config{
		EthHTTPURLs: "http://localhost:8545",
		EthWSURLs:   "ws://localhost:8546",
	}
}
