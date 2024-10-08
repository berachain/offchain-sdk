package eth_test

import (
	"bytes"
	"io"
	"os"
	"testing"

	"github.com/berachain/offchain-sdk/client/eth"
	"github.com/berachain/offchain-sdk/log"
	"github.com/stretchr/testify/require"
)

var (
	HTTPURL = os.Getenv("ETH_HTTP_URL")
	WSURL   = os.Getenv("ETH_WS_URL")
)

/******************************* HELPER FUNCTIONS ***************************************/

// NOTE: requires chain rpc url at env var `ETH_HTTP_URL` and `ETH_WS_URL`.
func checkEnv(t *testing.T) {
	ethHTTPRPC := os.Getenv("ETH_HTTP_URL")
	ethWSRPC := os.Getenv("ETH_WS_URL")
	if ethHTTPRPC == "" || ethWSRPC == "" {
		t.Skipf("Skipping test: no eth rpc url provided")
	}
}

// initConnectionPool initializes a new connection pool.
func initConnectionPool(
	cfg eth.ConnectionPoolConfig, writer io.Writer,
) (eth.ConnectionPool, error) {
	logger := log.NewLogger(writer, "test-runner")
	return eth.NewConnectionPoolImpl(cfg, logger)
}

// Use Init function as a setup function for the tests.
// It means each test will have to call Init function to set up the test.
func Init(
	cfg eth.ConnectionPoolConfig, writer io.Writer, t *testing.T,
) (eth.ConnectionPool, error) {
	checkEnv(t)
	return initConnectionPool(cfg, writer)
}

/******************************* TEST CASES ***************************************/

// TestNewConnectionPoolImpl_MissingURLs tests the case when the URLs are missing.
func TestNewConnectionPoolImpl_MissingURLs(t *testing.T) {
	cfg := eth.ConnectionPoolConfig{}
	var logBuffer bytes.Buffer

	_, err := Init(cfg, &logBuffer, t)
	require.ErrorContains(t, err, "ConnectionPool: missing URL for HTTP clients")
}

// TestNewConnectionPoolImpl_MissingWSURLs tests the case when the WS URLs are missing.
func TestNewConnectionPoolImpl_MissingWSURLs(t *testing.T) {
	cfg := eth.ConnectionPoolConfig{
		EthHTTPURLs: []string{HTTPURL},
	}
	var logBuffer bytes.Buffer
	pool, err := Init(cfg, &logBuffer, t)

	require.NoError(t, err)
	require.NotNil(t, pool)
	require.Contains(t, logBuffer.String(), "ConnectionPool: missing URL for WS clients")
}

// TestNewConnectionPoolImpl tests the case when the URLs are provided.
// It should the expected behavior.
func TestNewConnectionPoolImpl(t *testing.T) {
	cfg := eth.ConnectionPoolConfig{
		EthHTTPURLs: []string{HTTPURL},
		EthWSURLs:   []string{WSURL},
	}
	var logBuffer bytes.Buffer
	pool, err := Init(cfg, &logBuffer, t)

	require.NoError(t, err)
	require.NotNil(t, pool)
	require.Empty(t, logBuffer.String())
}

// TestGetHTTP tests the retrieval of the HTTP client when it
// has been set and the connection has been established.
func TestGetHTTP(t *testing.T) {
	cfg := eth.ConnectionPoolConfig{
		EthHTTPURLs: []string{HTTPURL},
	}
	var logBuffer bytes.Buffer
	pool, _ := Init(cfg, &logBuffer, t)
	err := pool.Dial("")
	require.NoError(t, err)

	client, ok := pool.GetHTTP()
	require.True(t, ok)
	require.NotNil(t, client)
}

// TestGetWS tests the retrieval of the HTTP client when it
// has been set and the connection has been established.
func TestGetWS(t *testing.T) {
	cfg := eth.ConnectionPoolConfig{
		EthHTTPURLs: []string{HTTPURL},
		EthWSURLs:   []string{WSURL},
	}
	var logBuffer bytes.Buffer
	pool, _ := Init(cfg, &logBuffer, t)
	err := pool.Dial("")

	require.NoError(t, err)

	client, ok := pool.GetWS()
	require.True(t, ok)
	require.NotNil(t, client)
}

// TestGetWS_WhenItIsNotSet tests the retrieval of the WS client when
// no WS URLs have been provided.
func TestGetWS_WhenItIsNotSet(t *testing.T) {
	cfg := eth.ConnectionPoolConfig{
		EthHTTPURLs: []string{HTTPURL},
	}
	var logBuffer bytes.Buffer
	pool, _ := Init(cfg, &logBuffer, t)
	err := pool.Dial("")
	require.NoError(t, err)

	client, ok := pool.GetWS()
	require.False(t, ok)
	require.Nil(t, client)
}
