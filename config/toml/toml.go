package toml

import (
	"strings"

	"github.com/spf13/viper"
)

// LoadConfig loads a TOML config file into the target. It will prioritize
// environment variables if envOverride is true.
func LoadConfig[T any](filepath string, target *T, envOverride bool, envPrefix string) error {
	// Find and read the config file
	viper.SetConfigFile(filepath)

	if envOverride {
		// Enable viper to read Environment Variables
		viper.AutomaticEnv()
		viper.SetEnvPrefix(envPrefix)
		viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	}

	if err := viper.ReadInConfig(); err != nil {
		return err
	}

	return viper.Unmarshal(target)
}
