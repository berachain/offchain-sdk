package eth

// Config is the configuration for the eth client.
type Config struct {
	EthHTTPURL string
	EthWSURL   string
}

func DefaultConfig() *Config {
	return &Config{
		EthHTTPURL: "http://localhost:8545",
		EthWSURL:   "ws://localhost:8546",
	}
}
