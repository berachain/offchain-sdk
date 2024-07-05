package eth

import (
	"context"
	"errors"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rpc"

	"github.com/berachain/offchain-sdk/telemetry"
)

// ErrClientNotFound is an error that is returned when a client is not found.
var (
	ErrClientNotFound = errors.New("client not found")
)

// ChainProviderImpl is an implementation of the ChainProvider interface.
type ChainProviderImpl struct {
	ConnectionPool
	rpcTimeout time.Duration
	m          telemetry.Metrics
}

// NewChainProviderImpl creates a new ChainProviderImpl with the given ConnectionPool.
func NewChainProviderImpl(pool ConnectionPool, cfg ConnectionPoolConfig) (Client, error) {
	c := &ChainProviderImpl{ConnectionPool: pool, rpcTimeout: cfg.DefaultTimeout}
	if c.rpcTimeout == 0 {
		c.rpcTimeout = defaultRPCTimeout
	}
	return c, nil
}

func (c *ChainProviderImpl) EnableMetrics(m telemetry.Metrics) {
	c.m = m
}

func (c *ChainProviderImpl) recordRPCMethod(
	rpcID, method string,
	start time.Time,
	err error, //nolint:unparam // false positive
) {
	if c.m == nil {
		return
	}
	tagsMap := map[string]string{"rpc_id": rpcID, "rpc_method": method}
	tags := formatTags(tagsMap)

	c.m.IncMonotonic("rpc.request.count", tags...)
	c.m.Time("rpc.request.duration", time.Since(start), tags...)

	if err != nil {
		c.m.IncMonotonic("rpc.request.error", tags...)
	}
}

// ==================================================================
// Implementations of Reader and Writer
// ==================================================================

// BlockByNumber returns the block for the given number.
func (c *ChainProviderImpl) BlockByNumber(
	ctx context.Context, num *big.Int) (*types.Block, error) {
	if client, ok := c.GetHTTP(); ok {
		ctxWithTimeout, cancel := context.WithTimeout(ctx, c.rpcTimeout)
		defer cancel()

		var err error
		defer c.recordRPCMethod(client.ClientID, "eth_getBlockByNumber", time.Now(), err)
		result, err := client.BlockByNumber(ctxWithTimeout, num)
		return result, err
	}
	return nil, ErrClientNotFound
}

// BlockReceipts returns the receipts for the given block number or hash.
func (c *ChainProviderImpl) BlockReceipts(
	ctx context.Context, blockNrOrHash rpc.BlockNumberOrHash) ([]*types.Receipt, error) {
	if client, ok := c.GetHTTP(); ok {
		ctxWithTimeout, cancel := context.WithTimeout(ctx, c.rpcTimeout)
		defer cancel()

		var err error
		defer c.recordRPCMethod(client.ClientID, "eth_getBlockReceipts", time.Now(), err)
		result, err := client.BlockReceipts(ctxWithTimeout, blockNrOrHash)
		return result, err
	}
	return nil, ErrClientNotFound
}

// TransactionReceipt returns the receipt for the given transaction hash.
func (c *ChainProviderImpl) TransactionReceipt(
	ctx context.Context, txHash common.Hash) (*types.Receipt, error) {
	if client, ok := c.GetHTTP(); ok {
		ctxWithTimeout, cancel := context.WithTimeout(ctx, c.rpcTimeout)
		defer cancel()

		var err error
		defer c.recordRPCMethod(client.ClientID, "eth_getTransactionReceipt", time.Now(), err)
		result, err := client.TransactionReceipt(ctxWithTimeout, txHash)
		return result, err
	}
	return nil, ErrClientNotFound
}

// SubscribeNewHead subscribes to new head events.
func (c *ChainProviderImpl) SubscribeNewHead(
	ctx context.Context) (chan *types.Header, ethereum.Subscription, error) {
	if client, ok := c.GetWS(); ok {
		ctxWithTimeout, cancel := context.WithTimeout(ctx, c.rpcTimeout)
		defer cancel()

		var err error
		defer c.recordRPCMethod(client.ClientID, "eth_subscribe", time.Now(), err)
		header, sub, err := client.SubscribeNewHead(ctxWithTimeout)
		return header, sub, err
	}
	return nil, nil, ErrClientNotFound
}

// BlockNumber returns the current block number.
func (c *ChainProviderImpl) BlockNumber(ctx context.Context) (uint64, error) {
	if client, ok := c.GetHTTP(); ok {
		ctxWithTimeout, cancel := context.WithTimeout(ctx, c.rpcTimeout)
		defer cancel()

		var err error
		defer c.recordRPCMethod(client.ClientID, "eth_blockNumber", time.Now(), err)
		result, err := client.BlockNumber(ctxWithTimeout)
		return result, err
	}
	return 0, ErrClientNotFound
}

// ChainID returns the current chain ID.
func (c *ChainProviderImpl) ChainID(ctx context.Context) (*big.Int, error) {
	if client, ok := c.GetHTTP(); ok {
		ctxWithTimeout, cancel := context.WithTimeout(ctx, c.rpcTimeout)
		defer cancel()

		var err error
		defer c.recordRPCMethod(client.ClientID, "eth_chainId", time.Now(), err)
		result, err := client.ChainID(ctxWithTimeout)
		return result, err
	}
	return nil, ErrClientNotFound
}

// BalanceAt returns the balance of the given address at the given block number.
func (c *ChainProviderImpl) BalanceAt(
	ctx context.Context, address common.Address, blockNumber *big.Int) (*big.Int, error) {
	if client, ok := c.GetHTTP(); ok {
		ctxWithTimeout, cancel := context.WithTimeout(ctx, c.rpcTimeout)
		defer cancel()

		var err error
		defer c.recordRPCMethod(client.ClientID, "eth_getBalance", time.Now(), err)
		result, err := client.BalanceAt(ctxWithTimeout, address, blockNumber)
		return result, err
	}
	return nil, ErrClientNotFound
}

// CodeAt returns the code of the given account at the given block number.
func (c *ChainProviderImpl) CodeAt(
	ctx context.Context, account common.Address, blockNumber *big.Int) ([]byte, error) {
	if client, ok := c.GetHTTP(); ok {
		ctxWithTimeout, cancel := context.WithTimeout(ctx, c.rpcTimeout)
		defer cancel()

		var err error
		defer c.recordRPCMethod(client.ClientID, "eth_getCode", time.Now(), err)
		result, err := client.CodeAt(ctxWithTimeout, account, blockNumber)
		return result, err
	}
	return nil, ErrClientNotFound
}

// EstimateGas estimates the gas needed to execute a specific transaction.
func (c *ChainProviderImpl) EstimateGas(
	ctx context.Context, msg ethereum.CallMsg) (uint64, error) {
	if client, ok := c.GetHTTP(); ok {
		ctxWithTimeout, cancel := context.WithTimeout(ctx, c.rpcTimeout)
		defer cancel()
		return client.EstimateGas(ctxWithTimeout, msg)
	}
	return 0, ErrClientNotFound
}

// FilterLogs returns the logs that satisfy the given filter query.
func (c *ChainProviderImpl) FilterLogs(
	ctx context.Context, q ethereum.FilterQuery) ([]types.Log, error) {
	if client, ok := c.GetHTTP(); ok {
		ctxWithTimeout, cancel := context.WithTimeout(ctx, c.rpcTimeout)
		defer cancel()

		var err error
		defer c.recordRPCMethod(client.ClientID, "eth_getLogs", time.Now(), err)
		result, err := client.FilterLogs(ctxWithTimeout, q)
		return result, err
	}
	return nil, ErrClientNotFound
}

// HeaderByNumber returns the header of the block with the given number.
func (c *ChainProviderImpl) HeaderByNumber(
	ctx context.Context, number *big.Int) (*types.Header, error) {
	if client, ok := c.GetHTTP(); ok {
		ctxWithTimeout, cancel := context.WithTimeout(ctx, c.rpcTimeout)
		defer cancel()

		var err error
		defer c.recordRPCMethod(client.ClientID, "eth_getBlockByNumber", time.Now(), err)
		result, err := client.HeaderByNumber(ctxWithTimeout, number)
		return result, err
	}
	return nil, ErrClientNotFound
}

// PendingCodeAt returns the code of the given account in the pending state.
func (c *ChainProviderImpl) PendingCodeAt(
	ctx context.Context, account common.Address) ([]byte, error) {
	if client, ok := c.GetHTTP(); ok {
		ctxWithTimeout, cancel := context.WithTimeout(ctx, c.rpcTimeout)
		defer cancel()

		var err error
		defer c.recordRPCMethod(client.ClientID, "eth_getCode", time.Now(), err)
		result, err := client.PendingCodeAt(ctxWithTimeout, account)
		return result, err
	}
	return nil, ErrClientNotFound
}

// PendingNonceAt returns the nonce of the given account in the pending state.
func (c *ChainProviderImpl) PendingNonceAt(
	ctx context.Context, account common.Address) (uint64, error) {
	if client, ok := c.GetHTTP(); ok {
		ctxWithTimeout, cancel := context.WithTimeout(ctx, c.rpcTimeout)
		defer cancel()

		var err error
		defer c.recordRPCMethod(client.ClientID, "eth_getTransactionCount", time.Now(), err)
		result, err := client.PendingNonceAt(ctxWithTimeout, account)
		return result, err
	}
	return 0, ErrClientNotFound
}

// PendingNonceAt returns the nonce of the given account in the pending state.
func (c *ChainProviderImpl) NonceAt(
	ctx context.Context, account common.Address, bn *big.Int) (uint64, error) {
	if client, ok := c.GetHTTP(); ok {
		ctxWithTimeout, cancel := context.WithTimeout(ctx, c.rpcTimeout)
		defer cancel()

		var err error
		defer c.recordRPCMethod(client.ClientID, "eth_getTransactionCount", time.Now(), err)
		result, err := client.NonceAt(ctxWithTimeout, account, bn)
		return result, err
	}
	return 0, ErrClientNotFound
}

// SendTransaction sends the given transaction.
func (c *ChainProviderImpl) SendTransaction(
	ctx context.Context, tx *types.Transaction) error {
	if client, ok := c.GetHTTP(); ok {
		ctxWithTimeout, cancel := context.WithTimeout(ctx, c.rpcTimeout)
		defer cancel()

		var err error
		defer c.recordRPCMethod(client.ClientID, "eth_sendTransaction", time.Now(), err)
		err = client.SendTransaction(ctxWithTimeout, tx)
		return err
	}
	return ErrClientNotFound
}

// SubscribeFilterLogs subscribes to new log events that satisfy the given filter query.
func (c *ChainProviderImpl) SubscribeFilterLogs(
	ctx context.Context, q ethereum.FilterQuery, ch chan<- types.Log,
) (ethereum.Subscription, error) {
	if client, ok := c.GetWS(); ok {
		ctxWithTimeout, cancel := context.WithTimeout(ctx, c.rpcTimeout)
		defer cancel()

		var err error
		defer c.recordRPCMethod(client.ClientID, "eth_subscribe", time.Now(), err)
		result, err := client.SubscribeFilterLogs(ctxWithTimeout, q, ch)
		return result, err
	}
	return nil, ErrClientNotFound
}

// SuggestGasPrice suggests a gas price.
func (c *ChainProviderImpl) SuggestGasPrice(ctx context.Context) (*big.Int, error) {
	if client, ok := c.GetHTTP(); ok {
		ctxWithTimeout, cancel := context.WithTimeout(ctx, c.rpcTimeout)
		defer cancel()

		var err error
		defer c.recordRPCMethod(client.ClientID, "eth_gasPrice", time.Now(), err)
		result, err := client.SuggestGasPrice(ctxWithTimeout)
		return result, err
	}
	return nil, ErrClientNotFound
}

// CallContract calls a contract with the given message at the given block number.
func (c *ChainProviderImpl) CallContract(
	ctx context.Context, msg ethereum.CallMsg, blockNumber *big.Int,
) ([]byte, error) {
	if client, ok := c.GetHTTP(); ok {
		ctxWithTimeout, cancel := context.WithTimeout(ctx, c.rpcTimeout)
		defer cancel()

		var err error
		defer c.recordRPCMethod(client.ClientID, "eth_call", time.Now(), err)
		result, err := client.CallContract(ctxWithTimeout, msg, blockNumber)
		return result, err
	}
	return nil, ErrClientNotFound
}

// SuggestGasTipCap suggests a gas tip cap.
func (c *ChainProviderImpl) SuggestGasTipCap(ctx context.Context) (*big.Int, error) {
	if client, ok := c.GetHTTP(); ok {
		ctxWithTimeout, cancel := context.WithTimeout(ctx, c.rpcTimeout)
		defer cancel()

		var err error
		defer c.recordRPCMethod(client.ClientID, "eth_gasTipCap", time.Now(), err)
		result, err := client.SuggestGasTipCap(ctxWithTimeout)
		return result, err
	}
	return nil, ErrClientNotFound
}

// TransactionByHash returns the transaction with the given hash.
func (c *ChainProviderImpl) TransactionByHash(
	ctx context.Context, hash common.Hash,
) (*types.Transaction, bool, error) {
	if client, ok := c.GetHTTP(); ok {
		ctxWithTimeout, cancel := context.WithTimeout(ctx, c.rpcTimeout)
		defer cancel()

		var err error
		defer c.recordRPCMethod(client.ClientID, "eth_getTransactionByHash", time.Now(), err)
		tx, pending, err := client.TransactionByHash(ctxWithTimeout, hash)
		return tx, pending, err
	}
	return nil, false, ErrClientNotFound
}

/*
TxPoolContentFrom returns the pending and queued transactions of this address.
Example response:

	{
		"pending": {
			1: {
				// transaction details...
			},
			...
		},
		"queued": {
			3: {
				// transaction details...
			},
			...
		}
	}
*/
func (c *ChainProviderImpl) TxPoolContentFrom(ctx context.Context, address common.Address) (
	map[string]map[uint64]*types.Transaction, error,
) {
	if client, ok := c.GetHTTP(); ok {
		ctxWithTimeout, cancel := context.WithTimeout(ctx, c.rpcTimeout)
		defer cancel()

		var err error
		defer c.recordRPCMethod(client.ClientID, "txpool_content", time.Now(), err)
		result, err := client.TxPoolContentFrom(ctxWithTimeout, address)
		return result, err
	}
	return nil, ErrClientNotFound
}

/*
TxPoolInspect returns the textual summary of all pending and queued transactions.
Example response:

	{
		"pending": {
			"0x12345": {
				1: "0x12345789: 1 wei + 2 gas x 3 wei"
			},
			...
		},
		"queued": {
			"0x67890": {
				2: "0x12345789: 1 wei + 2 gas x 3 wei"
			},
			...
		}
	}
*/
func (c *ChainProviderImpl) TxPoolInspect(ctx context.Context) (
	map[string]map[common.Address]map[uint64]string, error,
) {
	if client, ok := c.GetHTTP(); ok {
		ctxWithTimeout, cancel := context.WithTimeout(ctx, c.rpcTimeout)
		defer cancel()

		var err error
		defer c.recordRPCMethod(client.ClientID, "txpool_inspect", time.Now(), err)
		result, err := client.TxPoolInspect(ctxWithTimeout)
		return result, err
	}
	return nil, ErrClientNotFound
}

func (c *ChainProviderImpl) Health() bool {
	httpOk, wsOk := false, false
	if client, ok := c.GetHTTP(); ok {
		httpOk = client.Healthy()
	}
	if client, ok := c.GetWS(); ok {
		wsOk = client.Healthy()
	}
	return httpOk && wsOk
}
