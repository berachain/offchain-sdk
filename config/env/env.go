package env

import (
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
)

const (
	// Ethereum RPC URLs
	EnvEthRPCURL   = "ETH_RPC_URL"
	EnvEthWSURL    = "ETH_WS_URL"
	EnvEthRPCURLWS = "ETH_RPC_URL_WS" // Alternative WS URL used in some tests

	// Event listening configuration
	EnvEventName     = "EVENT_NAME"
	EnvAddressListen = "ADDRESS_TO_LISTEN"
)

// Loads environment variables from .env file
func Load() error {
	// Try loading from current directory first
	err := godotenv.Load()
	if err == nil {
		return nil
	}

	// Then If that fails, try to find .env in
	// parent directories
	dir, err := os.Getwd()
	if err != nil {
		return err
	}

	for {
		envPath := filepath.Join(dir, ".env")
		if _, err := os.Stat(envPath); err == nil {
			return godotenv.Load(envPath)
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	// If we get here, we couldn't find the .env file
	// But we don't return an error because the env vars
	// might be actually set in the system in which case
	// we don't need the .env file
	return nil
}

// Loads environment variables from the specified file
func LoadFile(filename string) error {
	return godotenv.Load(filename)
}

// Returns the Ethereum RPC URL
func GetEthRPCURL() string {
	return os.Getenv(EnvEthRPCURL)
}

// Returns the Ethereum WebSocket URL
func GetEthWSURL() string {
	if url := os.Getenv(EnvEthRPCURLWS); url != "" {
		return url
	}
	return os.Getenv(EnvEthWSURL)
}

// Returns the event name to listen for
func GetEventName() string {
	return os.Getenv(EnvEventName)
}

// Returns the contract address to listen to
func GetAddressToListen() string {
	return os.Getenv(EnvAddressListen)
}
