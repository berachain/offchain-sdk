package eth

import (
	"context"
	"sync"
	"time"

	"github.com/berachain/offchain-sdk/v2/log"

	"github.com/ethereum/go-ethereum/ethclient"
)

type HealthCheckedClient struct {
	*ExtendedEthClient
	dialurl             string
	logger              log.Logger
	healthy             bool
	healthCheckInterval time.Duration
	mu                  sync.Mutex
	clientID            string
}

func NewHealthCheckedClient(
	healthCheckInterval time.Duration, logger log.Logger,
) *HealthCheckedClient {
	return &HealthCheckedClient{
		logger:              logger,
		healthCheckInterval: healthCheckInterval,
	}
}

func (c *HealthCheckedClient) DialContextWithTimeout(
	ctx context.Context, rawurl string, rpcTimeout time.Duration,
) error {
	ctxWithTimeout, cancel := context.WithTimeout(ctx, c.healthCheckInterval)
	defer cancel()
	ethClient, err := ethclient.DialContext(ctxWithTimeout, rawurl)
	if err != nil {
		return err
	}

	c.ExtendedEthClient = NewExtendedEthClient(ethClient, rpcTimeout)
	c.dialurl = rawurl
	c.clientID = trimProtocolAndPort(rawurl)

	go c.StartHealthCheck(ctx)

	return nil
}

func (c *HealthCheckedClient) ClientID() string {
	return c.clientID
}

func (c *HealthCheckedClient) Health() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.healthy
}

func (c *HealthCheckedClient) SetHealthy(healthy bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.healthy = healthy
}

func (c *HealthCheckedClient) StartHealthCheck(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			ctxWithTimeout, cancel := context.WithTimeout(ctx, c.rpcTimeout)
			_, err := c.ChainID(ctxWithTimeout)
			cancel()
			if err != nil {
				c.SetHealthy(false)
				c.logger.Error("eth client reporting unhealthy", "err", err, "url", c.dialurl)
			} else {
				c.SetHealthy(true)
				c.logger.Debug("eth client reporting healthy", "url", c.dialurl)
			}
		}
		time.Sleep(c.healthCheckInterval)
	}
}
