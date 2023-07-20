package eth

import (
	"log"

	"github.com/spf13/viper"
)

// Config is the configuration for the eth client.
type Config struct {
	EthHTTPURL string `mapstructure:"ETH_RPC_URL"`
	EthWSURL   string `mapstructure:"ETH_WS_URL"`
}

// LoadConfig loads the configuration from the config file.
func LoadConfig(filepath string) Config {
	if filepath == "" {
		viper.AddConfigPath("./")   // Set the folder where the configuration file resides
		viper.SetConfigName(".env") // Name of the configuration file
		viper.SetConfigType("env")  // Set config file type to "env"
	} else {
		viper.SetConfigFile(filepath)
	}
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
