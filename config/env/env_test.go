package env

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEnv(t *testing.T) {
	t.Run("test loading from .env file", func(t *testing.T) {
		// Creating a temporary .env file for this test
		dir := t.TempDir()
		envFile := filepath.Join(dir, ".env")
		err := os.WriteFile(envFile, []byte(`
ETH_RPC_URL=http://localhost:8545
ETH_WS_URL=ws://localhost:8546
ETH_RPC_URL_WS=ws://localhost:8547
EVENT_NAME=NumberChanged(uint256)
ADDRESS_TO_LISTEN=0x5793a71D3eF074f71dCC21216Dbfd5C0e780132c
`), 0644)
		require.NoError(t, err)

		// Loading the env file
		err = LoadFile(envFile)
		require.NoError(t, err)

		// Testing each getter
		require.Equal(t, "http://localhost:8545", GetEthRPCURL())
		require.Equal(t, "ws://localhost:8547", GetEthWSURL(), "should prefer ETH_RPC_URL_WS")

		// Clearing ETH_RPC_URL_WS and verifying fallback to ETH_WS_URL
		os.Unsetenv(EnvEthRPCURLWS)
		require.Equal(t, "ws://localhost:8546", GetEthWSURL(), "should fallback to ETH_WS_URL")

		require.Equal(t, "NumberChanged(uint256)", GetEventName())
		require.Equal(t, "0x5793a71D3eF074f71dCC21216Dbfd5C0e780132c", GetAddressToListen())
	})

	t.Run("test loading non-existent file", func(t *testing.T) {
		err := LoadFile("non-existent.env")
		require.Error(t, err)
	})

	t.Run("test loading with missing values", func(t *testing.T) {
		// Clearing all env vars first
		os.Unsetenv(EnvEthRPCURL)
		os.Unsetenv(EnvEthWSURL)
		os.Unsetenv(EnvEthRPCURLWS)
		os.Unsetenv(EnvEventName)
		os.Unsetenv(EnvAddressListen)

		// Testing empty values
		require.Empty(t, GetEthRPCURL())
		require.Empty(t, GetEthWSURL())
		require.Empty(t, GetEventName())
		require.Empty(t, GetAddressToListen())
	})
}
