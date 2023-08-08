package config

// Config is the configuration for the eth client.
type Config struct {
	AddressToListen string `mapstructure:"ADDRESS_TO_LISTEN"`
	EventName       string `mapstructure:"EVENT_NAME"`
}
