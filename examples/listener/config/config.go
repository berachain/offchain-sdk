package config

import (
	"log"

	"github.com/spf13/viper"
)

// Config is the configuration for the eth client.
type Config struct {
	AddressToListen string `mapstructure:"ADDRESS_TO_LISTEN"`
	EventName       string `mapstructure:"EVENT_NAME"`
}

// LoadConfig loads the configuration from the config file.
func LoadConfig(filepath string) Config {
	viper.SetConfigFile(filepath)
	err := viper.ReadInConfig()
	if err != nil {
		log.Fatal("Error reading .env file:", err)
	}

	viper.AutomaticEnv()

	var config Config
	err = viper.Unmarshal(&config)
	if err != nil {
		log.Fatal("Error unmarshaling config:", err)
	}

	return config
}
