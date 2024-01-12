package eth

import (
	"context"
	"sync"
	"time"

	"github.com/berachain/offchain-sdk/log"

	"github.com/ethereum/go-ethereum/ethclient"
)

type HealthCheckedClient struct {
	*ExtendedEthClient
	dialurl             string
	logger              log.Logger
	healthy             bool
	healthCheckInterval time.Duration
	mu                  sync.Mutex
}

func NewHealthCheckedClient(logger log.Logger) *HealthCheckedClient {
	return &HealthCheckedClient{
		logger:              logger,
		healthCheckInterval: 5 * time.Second, //nolint:gomnd // todo paramaterize.
	}
}

func (c *HealthCheckedClient) DialContext(ctx context.Context, rawurl string) error {
	ethClient, err := ethclient.DialContext(ctx, rawurl)
	if err != nil {
		return err
	}
	c.ExtendedEthClient = &ExtendedEthClient{
		Client: ethClient,
	}

	c.dialurl = rawurl

	go c.StartHealthCheck(ctx)

	return nil
}

func (c *HealthCheckedClient) Healthy() bool {
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
			_, err := c.ChainID(ctx)
			if err != nil {
				c.SetHealthy(false)
				c.logger.Error("eth client reporting unhealthy", "err", err, "url", c.dialurl)
			} else {
				c.SetHealthy(true)
				c.logger.Info("eth client reporting healthy", "url", c.dialurl)
			}
		}
		time.Sleep(c.healthCheckInterval)
	}
}
