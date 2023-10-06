package toml

import (
	"strings"

	"github.com/spf13/viper"
)

// ReadTomlIntoMap reads a TOML file into a map.
func ReadIntoMap[T any](filepath string, target *T) error {
	return initConfig(filepath, false, "", target)
}

// PrioritizeEnv prioritizes environment variables over config file values.
func PrioritizeEnv[T any](filepath string, envPrefix string, target *T) error {
	return initConfig(filepath, true, envPrefix, target)
}

func initConfig[T any](filepath string, envOverride bool, envPrefix string, target *T) error {
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
