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
)

// ErrClientNotFound is an error that is returned when a client is not found.
var (
	ErrClientNotFound = errors.New("client not found")
)

// ChainProviderImpl is an implementation of the ChainProvider interface.
type ChainProviderImpl struct {
	ConnectionPool
	rpcTimeout time.Duration
}

// NewChainProviderImpl creates a new ChainProviderImpl with the given ConnectionPool.
func NewChainProviderImpl(pool ConnectionPool, cfg ConnectionPoolConfig) (Client, error) {
	c := &ChainProviderImpl{ConnectionPool: pool, rpcTimeout: cfg.DefaultTimeout}
	if c.rpcTimeout == 0 {
		c.rpcTimeout = defaultRPCTimeout
	}
	return c, nil
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
		return client.BlockByNumber(ctxWithTimeout, num)
	}
	return nil, ErrClientNotFound
}

// BlockReceipts returns the receipts for the given block number or hash.
func (c *ChainProviderImpl) BlockReceipts(
	ctx context.Context, blockNrOrHash rpc.BlockNumberOrHash) ([]*types.Receipt, error) {
	if client, ok := c.GetHTTP(); ok {
		ctxWithTimeout, cancel := context.WithTimeout(ctx, c.rpcTimeout)
		defer cancel()
		return client.BlockReceipts(ctxWithTimeout, blockNrOrHash)
	}
	return nil, ErrClientNotFound
}

// TransactionReceipt returns the receipt for the given transaction hash.
func (c *ChainProviderImpl) TransactionReceipt(
	ctx context.Context, txHash common.Hash) (*types.Receipt, error) {
	if client, ok := c.GetHTTP(); ok {
		ctxWithTimeout, cancel := context.WithTimeout(ctx, c.rpcTimeout)
		defer cancel()
		return client.TransactionReceipt(ctxWithTimeout, txHash)
	}
	return nil, ErrClientNotFound
}

// SubscribeNewHead subscribes to new head events.
func (c *ChainProviderImpl) SubscribeNewHead(
	ctx context.Context) (chan *types.Header, ethereum.Subscription, error) {
	if client, ok := c.GetWS(); ok {
		ctxWithTimeout, cancel := context.WithTimeout(ctx, c.rpcTimeout)
		defer cancel()
		return client.SubscribeNewHead(ctxWithTimeout)
	}
	return nil, nil, ErrClientNotFound
}

// BlockNumber returns the current block number.
func (c *ChainProviderImpl) BlockNumber(ctx context.Context) (uint64, error) {
	if client, ok := c.GetHTTP(); ok {
		ctxWithTimeout, cancel := context.WithTimeout(ctx, c.rpcTimeout)
		defer cancel()
		return client.BlockNumber(ctxWithTimeout)
	}
	return 0, ErrClientNotFound
}

// ChainID returns the current chain ID.
func (c *ChainProviderImpl) ChainID(ctx context.Context) (*big.Int, error) {
	if client, ok := c.GetHTTP(); ok {
		ctxWithTimeout, cancel := context.WithTimeout(ctx, c.rpcTimeout)
		defer cancel()
		return client.ChainID(ctxWithTimeout)
	}
	return nil, ErrClientNotFound
}

// BalanceAt returns the balance of the given address at the given block number.
func (c *ChainProviderImpl) BalanceAt(
	ctx context.Context, address common.Address, blockNumber *big.Int) (*big.Int, error) {
	if client, ok := c.GetHTTP(); ok {
		ctxWithTimeout, cancel := context.WithTimeout(ctx, c.rpcTimeout)
		defer cancel()
		return client.BalanceAt(ctxWithTimeout, address, blockNumber)
	}
	return nil, ErrClientNotFound
}

// CodeAt returns the code of the given account at the given block number.
func (c *ChainProviderImpl) CodeAt(
	ctx context.Context, account common.Address, blockNumber *big.Int) ([]byte, error) {
	if client, ok := c.GetHTTP(); ok {
		ctxWithTimeout, cancel := context.WithTimeout(ctx, c.rpcTimeout)
		defer cancel()
		return client.CodeAt(ctxWithTimeout, account, blockNumber)
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
		return client.FilterLogs(ctxWithTimeout, q)
	}
	return nil, ErrClientNotFound
}

// HeaderByNumber returns the header of the block with the given number.
func (c *ChainProviderImpl) HeaderByNumber(
	ctx context.Context, number *big.Int) (*types.Header, error) {
	if client, ok := c.GetHTTP(); ok {
		ctxWithTimeout, cancel := context.WithTimeout(ctx, c.rpcTimeout)
		defer cancel()
		return client.HeaderByNumber(ctxWithTimeout, number)
	}
	return nil, ErrClientNotFound
}

// PendingCodeAt returns the code of the given account in the pending state.
func (c *ChainProviderImpl) PendingCodeAt(
	ctx context.Context, account common.Address) ([]byte, error) {
	if client, ok := c.GetHTTP(); ok {
		ctxWithTimeout, cancel := context.WithTimeout(ctx, c.rpcTimeout)
		defer cancel()
		return client.PendingCodeAt(ctxWithTimeout, account)
	}
	return nil, ErrClientNotFound
}

// PendingNonceAt returns the nonce of the given account in the pending state.
func (c *ChainProviderImpl) PendingNonceAt(
	ctx context.Context, account common.Address) (uint64, error) {
	if client, ok := c.GetHTTP(); ok {
		ctxWithTimeout, cancel := context.WithTimeout(ctx, c.rpcTimeout)
		defer cancel()
		return client.PendingNonceAt(ctxWithTimeout, account)
	}
	return 0, ErrClientNotFound
}

// PendingNonceAt returns the nonce of the given account in the pending state.
func (c *ChainProviderImpl) NonceAt(
	ctx context.Context, account common.Address, bn *big.Int) (uint64, error) {
	if client, ok := c.GetHTTP(); ok {
		ctxWithTimeout, cancel := context.WithTimeout(ctx, c.rpcTimeout)
		defer cancel()
		return client.NonceAt(ctxWithTimeout, account, bn)
	}
	return 0, ErrClientNotFound
}

// SendTransaction sends the given transaction.
func (c *ChainProviderImpl) SendTransaction(
	ctx context.Context, tx *types.Transaction) error {
	if client, ok := c.GetHTTP(); ok {
		ctxWithTimeout, cancel := context.WithTimeout(ctx, c.rpcTimeout)
		defer cancel()
		return client.SendTransaction(ctxWithTimeout, tx)
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
		return client.SubscribeFilterLogs(ctxWithTimeout, q, ch)
	}
	return nil, ErrClientNotFound
}

// SuggestGasPrice suggests a gas price.
func (c *ChainProviderImpl) SuggestGasPrice(ctx context.Context) (*big.Int, error) {
	if client, ok := c.GetHTTP(); ok {
		ctxWithTimeout, cancel := context.WithTimeout(ctx, c.rpcTimeout)
		defer cancel()
		return client.SuggestGasPrice(ctxWithTimeout)
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
		return client.CallContract(ctxWithTimeout, msg, blockNumber)
	}
	return nil, ErrClientNotFound
}

// SuggestGasTipCap suggests a gas tip cap.
func (c *ChainProviderImpl) SuggestGasTipCap(ctx context.Context) (*big.Int, error) {
	if client, ok := c.GetHTTP(); ok {
		ctxWithTimeout, cancel := context.WithTimeout(ctx, c.rpcTimeout)
		defer cancel()
		return client.SuggestGasTipCap(ctxWithTimeout)
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
		return client.TransactionByHash(ctxWithTimeout, hash)
	}
	return nil, false, ErrClientNotFound
}

/*
TxPoolContentFrom returns the pending and queued transactions of this address.
Example response:

	{
		"pending": {
			"1": {
				// transaction details...
			},
			...
		},
		"queued": {
			"3": {
				// transaction details...
			},
			...
		}
	}
*/
func (c *ChainProviderImpl) TxPoolContentFrom(ctx context.Context, address common.Address) (
	map[string]map[string]*types.Transaction, error,
) {
	if client, ok := c.GetHTTP(); ok {
		ctxWithTimeout, cancel := context.WithTimeout(ctx, c.rpcTimeout)
		defer cancel()
		return client.TxPoolContentFrom(ctxWithTimeout, address)
	}
	return nil, ErrClientNotFound
}

/*
TxPoolInspect returns the textual summary of all pending and queued transactions.
Example response:

	{
		"pending": {
			"0x12345": {
				"1": "0x12345789: 1 wei + 2 gas x 3 wei"
			},
			...
		},
		"queued": {
			"0x67890": {
				"2": "0x12345789: 1 wei + 2 gas x 3 wei"
			},
			...
		}
	}
*/
func (c *ChainProviderImpl) TxPoolInspect(ctx context.Context) (
	map[string]map[common.Address]map[string]string, error,
) {
	if client, ok := c.GetHTTP(); ok {
		ctxWithTimeout, cancel := context.WithTimeout(ctx, c.rpcTimeout)
		defer cancel()
		return client.TxPoolInspect(ctxWithTimeout)
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
